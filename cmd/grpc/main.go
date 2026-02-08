package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fekuna/omnipos-pkg/database/postgres"
	"github.com/fekuna/omnipos-pkg/logger"
	storev1 "github.com/fekuna/omnipos-proto/proto/store/v1"
	"github.com/fekuna/omnipos-store-service/config"
	"github.com/fekuna/omnipos-store-service/internal/middleware"
	"github.com/fekuna/omnipos-store-service/internal/store/handler"
	"github.com/fekuna/omnipos-store-service/internal/store/repository"
	"github.com/fekuna/omnipos-store-service/internal/store/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	// Init Logger
	l := logger.NewZapLogger(&logger.ZapLoggerConfig{
		Level:         cfg.LoggerLvl,
		IsDevelopment: cfg.AppEnv == "development",
		Encoding:      "json", // Default
	})
	defer l.Sync()
	l.Info("Starting Store Service")

	// Init Database
	db, err := postgres.NewPostgres(&postgres.Config{
		Host:     cfg.PostgresHost,
		Port:     cfg.PostgresPort,
		User:     cfg.PostgresUser,
		Password: cfg.PostgresPass,
		DBName:   cfg.PostgresDB,
		SSLMode:  "disable",
	})
	if err != nil {
		l.Fatal(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	defer db.Close()

	// Verify DB
	if err := db.Ping(); err != nil {
		l.Fatal(fmt.Sprintf("Database ping failed: %v", err))
	}

	// Dependencies
	// Note: We use the *sqlx.DB from the postgres wrapper (db)
	repo := repository.NewPostgresRepository(db)
	uc := usecase.NewStoreUsecase(repo)
	h := handler.NewStoreHandler(uc, l)

	// Auth Interceptor
	authInterceptor := middleware.NewAuthContextInterceptor(l)

	// Start gRPC Server
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		l.Fatal(fmt.Sprintf("Failed to listen on port %s: %v", cfg.GRPCPort, err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)
	storev1.RegisterStoreServiceServer(grpcServer, h)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Graceful Shutdown
	go func() {
		l.Info(fmt.Sprintf("Store Service listening on port %s", cfg.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			l.Fatal(fmt.Sprintf("Failed to serve gRPC: %v", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down Store Service...")

	// Create a deadline to wait for.
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	l.Info("Store Service stopped")
}
