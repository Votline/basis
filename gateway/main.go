package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gateway/internal/routers"

	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewDevelopment()
	srv, err := routers.Init(log)
	if err != nil {
		log.Fatal("Failed to create server",
			zap.Error(err))
	}

	quit := make(chan os.Signal, 1)

	go func() {
		if err := srv.Start(); err != nil {
			log.Error("Failed to start server",
				zap.Error(err))
			quit <- os.Interrupt
		}
	}()

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	gracefulShutdown(srv, log)
}

func gracefulShutdown(srv *routers.Server, log *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("Shutting down server")
	if err := srv.Close(ctx); err != nil {
		log.Error("Error shutting down server", zap.Error(err))
	}
}
