package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/obalunenko/logger"
	"golang.org/x/sync/errgroup"

	"github.com/obalunenko/cthulhu-mythos-tools/internal/config"
	"github.com/obalunenko/cthulhu-mythos-tools/internal/service"
)

var errSignal = errors.New("received signal")

func main() {
	signals := make(chan os.Signal, 1)

	ctx := context.Background()

	cfg, err := config.Load(ctx)
	if err != nil {
		log.WithError(ctx, err).Fatal("Failed to load config")
	}

	l := log.Init(ctx, log.Params{
		Writer:     os.Stdout,
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		WithSource: false,
	})

	ctx = log.ContextWithLogger(ctx, l)

	printVersion(ctx)

	router := service.NewRouter()

	server := &http.Server{
		Addr:    net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port),
		Handler: router,
	}

	server.RegisterOnShutdown(func() {
		log.Info(ctx, "Server shutting down")

		server.SetKeepAlivesEnabled(false)

		log.Info(ctx, "Server shutdown complete")
	})

	ctx, cancel := context.WithCancelCause(ctx)
	defer func() {
		const msg = "Exit"

		var code int

		err := context.Cause(ctx)
		if err != nil && !errors.Is(err, errSignal) {
			code = 1
		}

		l := log.WithField(ctx, "cause", err)

		if code == 0 {
			l.Info(msg)

			return
		}

		l.Error(msg)

		os.Exit(code)
	}()

	defer cancel(nil)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	go func() {
		s := <-signals

		cancel(fmt.Errorf("%w: %s", errSignal, s.String()))
	}()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-ctx.Done()
		return server.Shutdown(ctx)
	})

	g.Go(func() error {
		log.WithFields(ctx, log.Fields{
			"address": server.Addr,
		}).Info("Server started")

		if err = server.ListenAndServe(); err != nil {
			if !errors.Is(http.ErrServerClosed, err) {
				log.WithError(ctx, err).Error("Failed to start server")

				return err
			}

			log.Info(ctx, "Server stopped gracefully")

			return nil
		}

		return nil
	})

	if err = g.Wait(); err != nil {
		log.WithError(ctx, err).Fatal("Service failed")
	}
}
