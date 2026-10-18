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
	"github.com/anand-gec/student-api/internal/http/handlers/student"
	"github.com/anand-gec/student-api/internal/storage/sqlite"
)

func main() {
	// load config
	cfg := config.MustLoad() // to import from others file package name than it's function name
	// database setup

	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("storage initialized", slog.String("env",cfg.Env),slog.String("version","1.0.0")) // used for versioning and also where to come value
	// setup router
	router := http.NewServeMux()

	//routers
	router.HandleFunc("POST /api/students", student.New(storage))
	router.HandleFunc("GET /api/students/{id}", student.GetById(storage))
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
