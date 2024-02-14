// TODO: cleanup
package ledger

import (
	"context"
	"strings"

	"github.com/deividaspetraitis/go/database/esdb"
	"github.com/deividaspetraitis/go/errors"
	"github.com/deividaspetraitis/go/es"
	"github.com/deividaspetraitis/go/log"

	esdbo "github.com/EventStore/EventStore-Client-Go/v3/esdb"
)

type GetWalletFunc func(ctx context.Context, id string) (*WalletAggregate, error)
type StoreWalletFunc func(ctx context.Context, w *WalletAggregate) error

func Subscription(ctx context.Context, getReadModel GetWalletFunc, storeReadModel StoreWalletFunc, getEventModel GetWalletFunc, sub *esdbo.Subscription) error {
	log.Println("Subscription", sub)

	iterator := esdb.NewSubscriptionIterator(sub)

	for iterator.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			event, err := iterator.Value()
			if err != nil {
				return err
			}

			if event.EventAppeared != nil {
				streamId := event.EventAppeared.OriginalEvent().StreamID
				walletID := strings.Split(streamId, "_")[1]
				revision := event.EventAppeared.OriginalEvent().EventNumber
				value, err := esdb.Value(event.EventAppeared)
				if err != nil {
				}

				log.Printf("received event %v@%v %s", revision, streamId, value.Data)

				// TODO: or from memory
				wallet, err := getReadModel(ctx, walletID)
				if err != nil && !errors.Is(err, ErrEntryNotFound) {
					return errors.Wrapf(err, "getReadModel: %#v", wallet)
				}

				log.Println(uint64(wallet.Version()-1), revision)
				if errors.Is(err, ErrEntryNotFound) || uint64(wallet.Version()-1) != revision {
					log.Printf("event %v@%v %s restored from events", revision, streamId, value.Data)
					wallet, err = getEventModel(ctx, walletID)
					if err != nil {
						log.WithError(err).Printf("getEventModel: %#v", wallet)
						continue
					}
				} else {
					log.Printf("event %v@%v %s applied", revision, streamId, value.Data)
					ev, err := es.GetAggregateEvent(&WalletAggregate{}, value.Type)
					if err != nil {
						log.WithError(err).Print("aggregate event not found")
						continue
					}

					if err := ev.UnmarshalJSON(value.Data); err != nil {
						return err
					}

					// apply new event
					wallet.Apply(es.NewEvent(value.AggregateID, &WalletAggregate{}, ev))
				}

				if err := storeReadModel(ctx, wallet); err != nil {
					return errors.Wrapf(err, "storeReadModel failure: %#v", wallet)
				}
			}

			if event.SubscriptionDropped != nil {
				log.Println("Subscription dropped", sub)
				break
			}
		}

	}

	return nil
}
