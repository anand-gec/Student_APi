package main

import (
	"context"
	"errors"
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
	// 1. Load config
	cfg := config.MustLoad()

	// 2. Setup JWT Secret
	jwt.SetSecret(cfg.JWTSecret)

	// 3. Database setup
	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	// Ensure database file locks/connections are released on shutdown
	defer func() {
		if closer, ok := any(storage).(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}()

	slog.Info("storage initialized", slog.String("env", cfg.Env), slog.String("version", "1.0.0"))

	// 4. Setup router
	router := http.NewServeMux()

	// Swagger docs
	router.HandleFunc("GET /swagger/", httpdocs.SwaggerUI)
	router.HandleFunc("GET /swagger/openapi.yaml", httpdocs.OpenAPISpec)

	// Admin endpoints
	router.HandleFunc("POST /api/admin/login", admin.Login(storage))
	router.HandleFunc("POST /api/admin/register", admin.Register(storage))
	router.HandleFunc("POST /api/admin/forgot-password", admin.ForgotPassword(storage))
	router.HandleFunc("POST /api/admin/reset-password", admin.ResetPassword(storage))

	// Protected student endpoints
	router.HandleFunc("POST /api/students", middleware.RequireAdminAuth(student.New(storage)))
	router.HandleFunc("PUT /api/students/{id}", middleware.RequireAdminAuth(student.UpdateStudent(storage)))
	router.HandleFunc("GET /api/students/{id}", middleware.RequireAdminAuth(student.GetById(storage)))
	router.HandleFunc("GET /api/students", middleware.RequireAdminAuth(student.GetList(storage)))
	router.HandleFunc("DELETE /api/students/{id}", middleware.RequireAdminAuth(student.DeleteStudentId(storage)))
	router.HandleFunc("DELETE /api/students", middleware.RequireAdminAuth(student.DeleteAll(storage)))

	// 5. Setup HTTP server with proper network timeouts
	server := http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}

	slog.Info("server started", slog.String("address", cfg.Addr))

	// Channel to listen for OS signals
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in background goroutine
	go func() {
		err := server.ListenAndServe()
		// FIX: Ignore http.ErrServerClosed because it is expected on graceful shutdown
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Block until signal is received
	<-done

	slog.Info("Shutting down the server...")

	// Context with timeout to give active requests time to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown server cleanly", slog.String("error", err.Error()))
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
