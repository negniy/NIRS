package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"watermark_service/internal/api"
	"watermark_service/internal/config"
	"watermark_service/internal/watermark"
)

func main() {
	cfg := config.Load()

	processor := watermark.NewProcessor(watermark.NoisePolicy{
		MinShiftRatio: cfg.MinShiftRatio,
		MaxShiftRatio: cfg.MaxShiftRatio,
	})

	handler := api.NewHandler(processor)
	router := api.NewRouter(handler)

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown error: %v", err)
		}
	}()

	log.Printf("watermark service listening on %s", cfg.Address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
