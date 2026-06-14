package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hugaojanuario/Paterna/internal/repository"
	"github.com/hugaojanuario/Paterna/pkg/bcrypt"
	"github.com/hugaojanuario/Paterna/pkg/session"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password required")
		return
	}

	user, err := repository.GetByEmail(req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if !bcrypt.CheckHash(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token := session.Create(user.ID, user.Email)

	writeJSON(w, http.StatusOK, loginResponse{
		Token: token,
		Email: user.Email,
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		writeError(w, http.StatusBadRequest, "missing Authorization header")
		return
	}

	token := strings.TrimPrefix(header, "Bearer ")
	session.Destroy(token)

	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}
