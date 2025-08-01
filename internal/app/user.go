package app

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/samrityy/go-oauth/internal/models"
)

func (a *App) HandleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := a.DB.Query("SELECT id, name, email FROM users")
		if err != nil {
			http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []models.User

		for rows.Next() {
			var u models.User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
				http.Error(w, "Failed to read user", http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}

		// Convert to JSON and write response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)

	case http.MethodPost:
		var u models.User
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		err = a.DB.QueryRow(
			"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
			u.Name, u.Email,
		).Scan(&u.ID)

		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(u)

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
		var u models.User
		err := a.DB.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id).
			Scan(&u.ID, &u.Name, &u.Email)
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)

	case http.MethodPatch:
		var u models.User
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Optional: You may validate fields before updating
		result, err := a.DB.Exec(
			"UPDATE users SET name = $1, email = $2 WHERE id = $3",
			u.Name, u.Email, id,
		)
		if err != nil {
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User updated"))

	case http.MethodDelete:
		result, err := a.DB.Exec("DELETE FROM users WHERE id = $1", id)
		if err != nil {
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}
	
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
	
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User deleted"))
	
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
