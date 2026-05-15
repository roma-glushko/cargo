package cmd

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/urfave/cli/v2"

	"github.com/roma-glushko/cargo/internal/booking"
	"github.com/roma-glushko/cargo/internal/cargo"
	"
	"github.com/roma-glushko/cargo/internal/eventbus"
	"github.com/roma-glushko/cargo/internal/handling"
	"github.com/roma-glushko/cargo/internal/inmem"
	"github.com/roma-glushko/cargo/internal/cargo"
	"github.com/roma-glushko/cargo/internal/pathfinder"
	"github.com/roma-glushko/cargo/internal/sample"
	"github.com/roma-glushko/cargo/internal/server"
)

func Run(cCtx *cli.Context) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cargoRepo := inmem.NewCargoRepository()
	locationRepo := inmem.NewLocationRepository()
	voyageRepo := inmem.NewVoyageRepository()
	eventRepo := inmem.NewHandlingEventRepository()

	ctx := cCtx.Context

	if err := sample.Populate(ctx, locationRepo, voyageRepo, cargoRepo, eventRepo); err != nil {
		return err
	}

	logger.Info("seed data loaded")

	bus := eventbus.New(100, logger)

	routingService := pathfinder.NewRoutingService(locationRepo, voyageRepo)
	bookingService := booking.NewService(cargoRepo, locationRepo, routingService)
	handlingService := handling.NewService(eventRepo, cargoRepo, locationRepo, voyageRepo, bus)
	inspectionService := inspection.NewService(cargoRepo, eventRepo, bus, logger)

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

	bookingHandler := server.NewBookingHandler(bookingService, cargoRepo, locationRepo)
	trackingHandler := server.NewTrackingHandler(cargoRepo, eventRepo)
	handlingHandler := server.NewHandlingHandler(handlingService)

	srv := server.New(bookingHandler, trackingHandler, handlingHandler)

	addr := ":8080"
	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv,
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		logger.Info("starting server", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	<-signalCtx.Done()
	logger.Info("shutting down")

	return httpServer.Shutdown(context.Background())
}
