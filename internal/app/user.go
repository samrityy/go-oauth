package app

import (
	"net/http"
	"strings"
)

func (a *App) HandleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List all users logic (example)
		w.Write([]byte("Get all users"))
	case http.MethodPost:
		// Create user logic (example)
		w.Write([]byte("Create user"))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func (a *App) HandleUserByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" {
		http.Error(w, "User ID missing", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Get user by ID: " + id))
	case http.MethodPatch:
		w.Write([]byte("Update user with ID: " + id))
	case http.MethodDelete:
		w.Write([]byte("Delete user with ID: " + id))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
