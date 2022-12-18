package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"

	"bb-project/db"
	"bb-project/internal/api"
	"bb-project/internal/config"
	"bb-project/internal/service"
	"bb-project/internal/storage"
	"bb-project/kafka"
)

func main() {
	cfg := config.InitConfig()

	// Init logger
	InitLogger(cfg.LogLevel, cfg.LogPretty)

	// Init DB
	db, err := db.InitConnection(cfg.Postgres.DSN, cfg.Postgres.Debug)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	dataPort := storage.NewDataPort(db)

	// Init Kafka Producer
	producer, err := kafka.NewProducer(cfg.Kafka.Host, cfg.Kafka.Topic)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	// Init the services
	callbackService := service.NewCallback(producer.Produce)
	objectService := service.NewObjectHandler(dataPort, cfg.ObjectEndpoint)

	// Init Kafka Consumer
	consumer, err := kafka.NewConsumer(cfg.Kafka.Host, cfg.Kafka.GroupId, []string{cfg.Kafka.Topic}, "earliest")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	consumer.Consume(objectService.Handle)

	// Init ClearUp service
	clearUp := service.NewClearUp(dataPort)
	clearUp.Run()

	// Init a router
	e := api.NewRouter(callbackService)
	// Start server
	log.Info().Msgf("Start service on http://%s", cfg.ApiListener)
	go func() {
		if err := e.Start(cfg.ApiListener); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("Service - listen: %s", err.Error())
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error")
	}
	clearUp.Stop()
	consumer.Stop()
	producer.Stop()
	db.Close()
}
