package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/urfave/cli/v2"

	"github.com/roma-glushko/cargo/internal/booking"
	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/eventbus"
	"github.com/roma-glushko/cargo/internal/handling"
	"github.com/roma-glushko/cargo/internal/inmem"
	"github.com/roma-glushko/cargo/internal/inspection"
	"github.com/roma-glushko/cargo/internal/location"
	"github.com/roma-glushko/cargo/internal/pathfinder"
	"github.com/roma-glushko/cargo/internal/postgres"
	"github.com/roma-glushko/cargo/internal/sample"
	"github.com/roma-glushko/cargo/internal/server"
	"github.com/roma-glushko/cargo/internal/voyage"
)

var RunFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "store",
		Value:   "inmem",
		Usage:   "storage backend: inmem or postgres",
		EnvVars: []string{"CARGO_STORE"},
	},
	&cli.StringFlag{
		Name:    "db",
		Value:   "postgres://cargo:shipthatpackage@localhost:5432/cargo?sslmode=disable",
		Usage:   "PostgreSQL connection string",
		EnvVars: []string{"CARGO_DATABASE_URL"},
	},
	&cli.StringFlag{
		Name:    "addr",
		Value:   ":8080",
		Usage:   "HTTP listen address",
		EnvVars: []string{"CARGO_ADDR"},
	},
}

type repos struct {
	cargos    cargo.Repository
	locations location.Repository
	voyages   voyage.Repository
	events    cargo.HandlingEventRepository
}

func Run(cCtx *cli.Context) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := cCtx.Context

	r, cleanup, err := initRepos(ctx, cCtx, logger)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	if err := sample.Populate(ctx, r.locations, r.voyages, r.cargos, r.events); err != nil {
		return err
	}

	logger.Info("seed data loaded")

	bus := eventbus.New(100, logger)

	routingService := pathfinder.NewRoutingService(r.locations, r.voyages)
	bookingService := booking.NewService(r.cargos, r.locations, routingService)
	handlingService := handling.NewService(r.events, r.cargos, r.locations, r.voyages, bus)
	inspectionService := inspection.NewService(r.cargos, r.events, bus, logger)

	bus.Subscribe(eventbus.TopicCargoHandled, func(id cargo.TrackingID) {
		inspectionService.InspectCargo(id)
	})
	bus.Subscribe(eventbus.TopicCargoMisdirected, func(id cargo.TrackingID) {
		logger.Warn("cargo misdirected", "trackingId", string(id))
	})
	bus.Subscribe(eventbus.TopicCargoArrived, func(id cargo.TrackingID) {
		logger.Info("cargo arrived at destination", "trackingId", string(id))
	})

	bus.Start()
	defer bus.Stop()

	bookingHandler := server.NewBookingHandler(bookingService, r.cargos, r.locations)
	trackingHandler := server.NewTrackingHandler(r.cargos, r.events)
	handlingHandler := server.NewHandlingHandler(handlingService)

	srv := server.New(bookingHandler, trackingHandler, handlingHandler)

	addr := cCtx.String("addr")
	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv,
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		logger.Info("starting server", "addr", addr, "store", cCtx.String("store"))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	<-signalCtx.Done()
	logger.Info("shutting down")

	return httpServer.Shutdown(context.Background())
}

func initRepos(ctx context.Context, cCtx *cli.Context, logger *slog.Logger) (repos, func(), error) {
	switch cCtx.String("store") {
	case "inmem":
		return repos{
			cargos:    inmem.NewCargoRepository(),
			locations: inmem.NewLocationRepository(),
			voyages:   inmem.NewVoyageRepository(),
			events:    inmem.NewHandlingEventRepository(),
		}, nil, nil

	case "postgres":
		connString := cCtx.String("db")
		pool, err := postgres.NewPool(ctx, connString)
		if err != nil {
			return repos{}, nil, fmt.Errorf("connecting to postgres: %w", err)
		}

		logger.Info("connected to PostgreSQL", "connString", connString)

		if err := postgres.Migrate(ctx, pool); err != nil {
			pool.Close()
			return repos{}, nil, fmt.Errorf("running migrations: %w", err)
		}

		logger.Info("database migrations applied")

		return repos{
			cargos:    postgres.NewCargoRepository(pool),
			locations: postgres.NewLocationRepository(pool),
			voyages:   postgres.NewVoyageRepository(pool),
			events:    postgres.NewHandlingEventRepository(pool),
		}, pool.Close, nil

	default:
		return repos{}, nil, fmt.Errorf("unknown store: %s (use 'inmem' or 'postgres')", cCtx.String("store"))
	}
}
