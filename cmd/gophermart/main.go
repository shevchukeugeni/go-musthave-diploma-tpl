package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/shevchukeugeni/gofermart/internal/server"
	"github.com/shevchukeugeni/gofermart/internal/store/order"
	"github.com/shevchukeugeni/gofermart/internal/store/postgres"
	"github.com/shevchukeugeni/gofermart/internal/store/user"
	"github.com/shevchukeugeni/gofermart/internal/store/withdrawal"
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

	if envDBURL := os.Getenv("DATABASE_URI"); envDBURL != "" {
		dbURI = envDBURL
	}

	if asa := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); asa != "" {
		accrualSystemAddr = asa
	}
}

func main() {
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	db, err := postgres.NewPostgresDB(postgres.Config{URL: dbURI})
	if err != nil {
		logger.Fatal("failed to initialize db: " + err.Error())
	}
	defer db.Close()

	ctx, cancelCtx := context.WithCancel(context.Background())

	userRepo := user.NewRepository(db)
	orderRepo := order.NewRepository(db)
	withdrawalRepo := withdrawal.NewRepository(db, orderRepo)

	client := resty.New()

	updater := worker.NewWorker(logger, db, accrualSystemAddr, orderRepo, client)

	go updater.Run(ctx)

	router := server.SetupRouter(logger, userRepo, orderRepo, withdrawalRepo)

	logger.Info("Running server on", zap.String("address", flagRunAddr))
	err = http.ListenAndServe(flagRunAddr, router)
	if err != http.ErrServerClosed {
		logger.Fatal("HTTP server ListenAndServe Error", zap.Error(err))
	}

	cancelCtx()
}
