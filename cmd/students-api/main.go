package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anand-gec/student-api/internal/config"
)

func main() {
	// load config
	cfg := config.MustLoad()
	// database setup
	// setup router
	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to Student api"))
	})
	// setup server

	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	slog.Info("server started", slog.String("address", cfg.Addr))
	// fmt.Printf("Server Started  %s", cfg.HTTPServer.Addr)

	done := make(chan os.Signal, 1)

	//notify when interrupt signal get hare.
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// for graceful ShortDown the code.. using this functions
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("Failed to start Server")
		}
	}()

	<-done

	slog.Info(("Shutting Down the server"))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to Shutdown server ", slog.String("error", err.Error()))
	}

	/*
		// same as upper
			err := server.Shutdown(ctx)
		if err != nil {
			slog.Error("Failed to Shutdown server ", slog.String("error", err.Error()))
		}
	*/

	slog.Info("server shutdown successfully")
}
