package main

import (
	"damirqa/loyalty-system/internal/repository/pg"
	"damirqa/loyalty-system/internal/service"
	"database/sql"
	"flag"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"damirqa/loyalty-system/internal/config"
	delivery "damirqa/loyalty-system/internal/delivery/http"
)

func main() {
	runAddress := flag.String("a", os.Getenv("RUN_ADDRESS"), "Run address")
	dbURI := flag.String("d", os.Getenv("DATABASE_URI"), "Database URI")
	accrualAddress := flag.String("r", os.Getenv("ACCRUAL_SYSTEM_ADDRESS"), "Accrual system address")
	migrationsPath := flag.String("m", "./migrations", "Path to migration files")
	flag.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Config{
		RunAddress:           *runAddress,
		DatabaseURI:          *dbURI,
		AccrualSystemAddress: *accrualAddress,
	}

	applyMigrations(cfg.DatabaseURI, *migrationsPath)

	db, err := sql.Open("postgres", cfg.DatabaseURI)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	userRepo := pg.NewUserRepository(db)
	orderRepo := pg.NewOrderRepository(db)
	balanceRepo := pg.NewBalanceRepository(db)

	authService := service.NewAuthService(userRepo)
	orderService := service.NewOrderService(orderRepo)
	balanceService := service.NewBalanceService(balanceRepo, orderRepo)
	accrualClient := service.NewAccrualClient(cfg.AccrualSystemAddress)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	apiHandler := delivery.NewAPIHandler(authService, orderService, balanceService, accrualClient, logger)
	apiHandler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         cfg.RunAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Starting server", zap.String("address", cfg.RunAddress))
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func applyMigrations(databaseURL, migrationsPath string) {
	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Ошибка инициализации миграций: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка применения миграций: %v", err)
	}

	log.Println("Миграции успешно применены или уже актуальны")
}
