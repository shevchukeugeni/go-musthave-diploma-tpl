package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"go.uber.org/zap"

	"github.com/shevchukeugeni/gofermart/internal/server"
	"github.com/shevchukeugeni/gofermart/internal/store/postgres"
	"github.com/shevchukeugeni/gofermart/internal/worker"
)

var flagRunAddr, dbURI, accrualSystemAddr string

func init() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&dbURI, "d", "", "database connection uri")
	flag.StringVar(&accrualSystemAddr, "r", "", "accrual system address")

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envDBURL := os.Getenv("DATABASE_DSN"); envDBURL != "" {
		dbURI = envDBURL
	}

	if asa := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); asa != "" {
		accrualSystemAddr = asa
	}
}

func main() {
	//NOTE: Greetings to the inspector,
	//This version is a simple template that I plan to beef up in the coming days. Actually I am only interested in one thing here.
	//I've used middleware of chi library with token generation via lestratt-go library to check tokens.
	//Is it necessary to implement session storage and refresh token generation in this project?
	//In the metrics service that I performed during the first part of the training there was no authorization via jwt, so I'm clarifying.

	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	db, err := postgres.NewPostgresDB(postgres.Config{URL: dbURI})
	if err != nil {
		logger.Error("failed to initialize db: " + err.Error())
	}
	defer db.Close()

	ctx, cancelCtx := context.WithCancel(context.Background())

	updater := worker.NewWorker(logger, db, accrualSystemAddr)

	go updater.Run(ctx)

	router := server.SetupRouter(logger)

	logger.Info("Running server on", zap.String("address", flagRunAddr))
	err = http.ListenAndServe(flagRunAddr, router)
	if err != http.ErrServerClosed {
		logger.Fatal("HTTP server ListenAndServe Error", zap.Error(err))
	}

	cancelCtx()
}
