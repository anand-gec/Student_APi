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
	"github.com/anand-gec/student-api/internal/config/utils/jwt"
	httpdocs "github.com/anand-gec/student-api/internal/http"
	"github.com/anand-gec/student-api/internal/http/handlers/admin"
	"github.com/anand-gec/student-api/internal/http/handlers/student"
	"github.com/anand-gec/student-api/internal/http/middleware"
	"github.com/anand-gec/student-api/internal/storage/sqlite"
)

func main() {
	// load config
	cfg := config.MustLoad() // to import from others file package name than it's function name
	// database setup

	// JWT secret from config
	jwt.SetSecret(cfg.JWTSecret)

	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("storage initialized", slog.String("env", cfg.Env), slog.String("version", "1.0.0")) // used for versioning and also where to come value
	// setup router
	router := http.NewServeMux()

	router.HandleFunc("GET /swagger/", httpdocs.SwaggerUI)
	router.HandleFunc("GET /swagger/openapi.yaml", httpdocs.OpenAPISpec)

	// Admin login endpoint to obtain JWT token
	router.HandleFunc("POST /api/admin/login", admin.Login(storage))
	// Optional: Admin registration endpoint
	router.HandleFunc("POST /api/admin/register", admin.Register(storage))

	// All student actions wrapped with RequireAdminAuth middleware
	router.HandleFunc("POST /api/students", middleware.RequireAdminAuth(student.New(storage)))
	router.HandleFunc("GET /api/students/{id}", middleware.RequireAdminAuth(student.GetById(storage)))
	router.HandleFunc("GET /api/students", middleware.RequireAdminAuth(student.GetList(storage)))
	router.HandleFunc("DELETE /api/students/{id}", middleware.RequireAdminAuth(student.DeleteStudentId(storage)))
	router.HandleFunc("DELETE /api/students", middleware.RequireAdminAuth(student.DeleteAll(storage)))

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

		// Ignore ErrServerClosed during normal shutdown
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
			if err := server.Shutdown(ctx); err != nil {
			slog.Error("Failed to Shutdown server ", slog.String("error", err.Error()))
		}

	*/

	slog.Info("server shutdown successfully")
}
