package admin

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/anand-gec/student-api/internal/config/utils/jwt"
	"github.com/anand-gec/student-api/internal/storage"
	"github.com/anand-gec/student-api/internal/storage/sqlite"
	"github.com/anand-gec/student-api/internal/utils/response"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
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

func ForgotPassword(storage *sqlite.Sqlite) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ForgotPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
			response.WriteJson(w, http.StatusBadRequest, map[string]string{"error": "valid email is required"})
			return
		}
		_, err := storage.GetAdminByEmail(req.Email)
		if err != nil {
			// For security, you can return 200 to prevent email enumeration,
			// or 404/400 for explicit API responses:
			response.WriteJson(w, http.StatusNotFound, map[string]string{"error": "admin with this email does not exist"})
			return
		}

		token, err := storage.StoreResetToken(req.Email)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate reset token"})
			return
		}

		// In production, you would send this token via an SMTP email service.
		// For local API testing, return it in the response payload:
		response.WriteJson(w, http.StatusOK, map[string]any{
			"status":      "success",
			"message":     "password reset token generated (valid for 15 minutes)",
			"reset_token": token,
		})
	}
}

func ResetPassword(storage *sqlite.Sqlite) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ResetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJson(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
			return
		}

		if req.Token == "" || req.NewPassword == "" {
			response.WriteJson(w, http.StatusBadRequest, map[string]string{"error": "token and new_password are required"})
			return
		}

		if len(req.NewPassword) < 6 {
			response.WriteJson(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 6 characters"})
			return
		}

		err := storage.ResetPasswordWithToken(req.Token, req.NewPassword)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		response.WriteJson(w, http.StatusOK, map[string]string{
			"status":  "success",
			"message": "password reset successfully. You can now login with your new password.",
		})
	}
}
