package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/deividaspetraitis/ledger/config"
	ihttp "github.com/deividaspetraitis/ledger/http"

	"github.com/deividaspetraitis/go/database/esdb"
	"github.com/deividaspetraitis/go/errors"
	"github.com/deividaspetraitis/go/log"
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
	esclient, err := esdb.NewClient(cfg.Database)
	if err != nil {
		return errors.Wrap(err, "unable connect to database instance")
	}

	// =========================================================================
	// Start HTTP server

	api := http.Server{
		Addr:    cfg.HTTP.Address,
		Handler: ihttp.API(shutdown, cfg.HTTP, logger, esclient),
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

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			logger.WithError(err).Error("graceful shutdown did not complete")
			api.Close()
			esclient.Close()
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
