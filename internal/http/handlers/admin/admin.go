package admin

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/anand-gec/student-api/internal/config/utils/jwt"
	"github.com/anand-gec/student-api/internal/storage"
	"github.com/anand-gec/student-api/internal/utils/response"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("1. Hit Login handler")

		var req LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed decoding JSON payload", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
			return
		}

		slog.Info("2. Received payload", slog.String("email", req.Email))

		admin, err := storage.GetAdminByEmail(req.Email)
		if err != nil {
			slog.Error("3. Admin lookup failed in SQLite", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}

		slog.Info("4. Admin record found in database", slog.Int64("id", admin.Id))

		if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
			slog.Error("5. Bcrypt password comparison failed", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}

		slog.Info("6. Password matched successfully")

		token, err := jwt.GenerateToken(admin.Id, admin.Email)
		if err != nil {
			slog.Error("7. JWT generation failed", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
			return
		}

		slog.Info("8. Token issued successfully")

		response.WriteJson(w, http.StatusOK, map[string]any{
			"status": "success",
			"token":  token,
		})
	}
}

func Register(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJson(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
			return
		}

		if req.Email == "" || req.Password == "" {
			response.WriteJson(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
			return
		}

		id, err := storage.CreateAdmin(req.Email, req.Password)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		response.WriteJson(w, http.StatusCreated, map[string]any{
			"status":   "Success",
			"Admin_id": id,
		})
	}
}
