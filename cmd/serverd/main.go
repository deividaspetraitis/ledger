package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/deividaspetraitis/ledger"
	"github.com/deividaspetraitis/ledger/config"
	dbasynq "github.com/deividaspetraitis/ledger/database/asynq"
	eventdb "github.com/deividaspetraitis/ledger/database/esdb"
	readdb "github.com/deividaspetraitis/ledger/database/sql"
	ihttp "github.com/deividaspetraitis/ledger/http"

	"github.com/deividaspetraitis/go/database/esdb"
	"github.com/deividaspetraitis/go/database/sql"
	"github.com/deividaspetraitis/go/errors"
	"github.com/deividaspetraitis/go/es"
	"github.com/deividaspetraitis/go/log"

	esdbo "github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/hibiken/asynq"
)

var shutdowntimeout = time.Duration(5) * time.Second

// program flags
var (
	cfgPath    string
	cpuprofile string
)

// initialise program state
func init() {
	flag.StringVar(&cfgPath, "config", os.Getenv("config"), "PATH to .env configuration file")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
}

// main program entry point.
func main() {
	flag.Parse()

	logger := log.Default()
	ctx := context.Background()

	if len(cpuprofile) > 0 {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f) // nolint
		defer pprof.StopCPUProfile()
	}

	cfg, err := config.New(cfgPath)
	if err != nil {
		logger.WithError(err).Fatal("parsing configuration file")
	}

	if err := run(ctx, cfg, logger); err != nil {
		logger.WithError(err).Fatal("unable to start service")
	}
}

func run(ctx context.Context, cfg *config.Config, logger log.Logger) error {
	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// =========================================================================
	// Construct services

	// connect to DB instance
	esclient, err := esdb.NewClient(cfg.Database.EventStore)
	if err != nil {
		return errors.Wrap(err, "unable connect to EventStore instance")
	}

	sqldb := sql.NewDB(ctx, cfg.Database.Postgres)
	if err := sqldb.Open(); err != nil {
		return errors.Wrap(err, "unable connect to PostgreSQL instance")
	}

	sub, err := esclient.SubscribeToAll(ctx, esdbo.SubscribeToAllOptions{
		Filter: esdbo.ExcludeSystemEventsFilter(),
		// Filter: &esdbo.SubscriptionFilter{
		// 	Type:  esdbo.EventFilterType,
		// 	Regex: "^Wallet",
		// },
	})

	if err != nil {
		panic(err)
	}

	defer sub.Close()

	// TODO: uncomment
	// TODO: safety?

	go ledger.Subscription(ctx, func(ctx context.Context, id string) (*ledger.WalletAggregate, error) {
		return readdb.GetWallet(ctx, sqldb, id)
	}, func(ctx context.Context, w *ledger.WalletAggregate) error {
		return readdb.StoreWallet(ctx, sqldb, w)
	}, func(ctx context.Context, id string) (*ledger.WalletAggregate, error) {
		return eventdb.Get[*ledger.WalletAggregate](ctx, esclient, &ledger.WalletAggregate{}, id)
	}, sub)

	redisAddr := fmt.Sprintf("%s:%d", cfg.Database.Redis.Host, cfg.Database.Redis.Port)
	asynqc := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	asynqi := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisAddr})
	asynqsrv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)

	cache := ledger.NewWithCache[*ledger.WalletAggregate](ledger.NewInMemory[*ledger.WalletAggregate](), func(id string) (*ledger.WalletAggregate, error) {
		return ledger.GetWallet(ctx, func(ctx context.Context, aggregate es.Aggregate, id string) (*ledger.WalletAggregate, error) {
			return eventdb.Get[*ledger.WalletAggregate](ctx, esclient, aggregate, id)
		}, id)
	})

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.Handle(string(dbasynq.TaskCreateWallet), dbasynq.NewCreateWalletProcessor(logger, func(ctx context.Context, aggregate es.Aggregate) error {
		return eventdb.Save(ctx, esclient, aggregate)
	}))
	mux.Handle(string(dbasynq.TaskCreateTransaction), dbasynq.NewCreateTransactionProcessor(logger, func(ctx context.Context, aggregate es.Aggregate, id string) (*ledger.WalletAggregate, error) {
		return eventdb.Get[*ledger.WalletAggregate](ctx, esclient, aggregate, id)
	}, func(ctx context.Context, aggregate es.Aggregate) error {
		return eventdb.Save(ctx, esclient, aggregate)
	}))

	go func() {
		logger.Println("asynq server listening for new tasks")
		serverErrors <- asynqsrv.Run(mux)
	}()

	defer asynqi.Close()
	// asynqc.Enqueue(ctx)

	// =========================================================================
	// Start HTTP server

	api := http.Server{
		Addr:    cfg.HTTP.Address,
		Handler: ihttp.API(shutdown, cfg.HTTP, logger, esclient, asynqc, asynqi, cache),
	}

	go func() {
		logger.Printf("http server listening on %s", cfg.HTTP.Address)
		serverErrors <- api.ListenAndServe()
	}()

	// ========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		logger.Printf("http server start shutdown caused by %v", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(ctx, shutdowntimeout)
		defer cancel()

		if err := esclient.Close(); err != nil {
			logger.WithError(err).Error("graceful shutdown did not complete")
		}

		if err := asynqc.Close(); err != nil {
			logger.WithError(err).Error("graceful shutdown did not complete")
		}

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			logger.WithError(err).Error("graceful shutdown did not complete")
			api.Close()
			esclient.Close()
			asynqc.Close()
		}

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		case err != nil:
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}

	return nil
}
