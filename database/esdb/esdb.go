package eventstore

import (
	"context"

	"github.com/deividaspetraitis/ledger"

	"github.com/deividaspetraitis/go/database/esdb"
	"github.com/deividaspetraitis/go/errors"
	"github.com/deividaspetraitis/go/es"
	"github.com/deividaspetraitis/go/log"
)

// Save persists an aggregate into underlying DB store.
// Save calls aggregate.Sync if store operations were successful.
func Save(ctx context.Context, db *esdb.Client, aggregate es.Aggregate) error {
	var (
		events    []*esdb.Event
		processed []*es.Event
	)

	for _, v := range aggregate.Events() {
		bytes, err := v.Data.MarshalJSON()
		if err != nil {
			return errors.Wrap(err, "failed to serialise")
		}

		events = append(events, &esdb.Event{
			AggregateID: v.AggregateID,
			Version:     esdb.Version(v.Version),
			Aggregate:   es.ParseAggregateName(v.Aggregate),
			Type:        es.ParseEventName(v.Data),
			Timestamp:   v.Timestamp,
			Data:        bytes,
			Metadata:    v.Metadata,
		})

		processed = append(processed, v)
	}

	if err := db.Save(ctx, events); err != nil {
		return err
	}

	// mark recently stored events as processed.
	for _, v := range processed {
		if err := aggregate.Sync(v); err != nil {
			return err
		}
	}

	return nil
}

// Get retrieves aggregate with restored state from underlying DB store.
func Get[T any](ctx context.Context, db *esdb.Client, aggregate es.Aggregate, id string) (T, error) {
	iterator, err := db.Get(ctx, id, es.ParseAggregateName(aggregate), esdb.Version(aggregate.Root().Version()))
	if err != nil {
		return *new(T), err
	}
	defer iterator.Close()

	var events []*es.Event
	for iterator.Next() {
		select {
		case <-ctx.Done():
			return *new(T), ctx.Err()
		default:
			event, err := iterator.Value()
			if err != nil {
				return *new(T), err
			}

			ev, err := es.GetAggregateEvent(aggregate, event.Type)
			if err != nil {
				log.WithError(err).Print("aggregate event not found")
				continue
			}

			if err := ev.UnmarshalJSON(event.Data); err != nil {
				return *new(T), err
			}

			events = append(events, es.NewEvent(id, aggregate, ev))
		}
	}

	if err := iterator.Error(); err != nil {
		return *new(T), err
	}

	// reconstruct state
	if err := aggregate.Reply(events); err != nil {
		return *new(T), err
	}

	// no events for given aggregate were found
	// meaning such aggregate does not exit
	if aggregate.Root().Version() == 0 {
		return *new(T), ledger.ErrEntryNotFound
	}

	return aggregate.(T), nil
}
