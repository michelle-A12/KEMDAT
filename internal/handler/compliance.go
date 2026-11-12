package handler

import (
	"aaa_bekend/internal/db"
	"aaa_bekend/internal/lmodel"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func GetMyData(w http.ResponseWriter, r *http.Request) {
	user, _ := lmodel.UserFromContext(r.Context())

	profile, err := db.GetUserByID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(profile)
}

func UpdateMyData(w http.ResponseWriter, r *http.Request) {
	user, _ := lmodel.UserFromContext(r.Context())

	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		http.Error(w, "Username must not be empty", http.StatusBadRequest)
		return
	}

	if err := db.UpdateMyData(r.Context(), user.ID, req.Username); err != nil {
		// kemungkinan besar username sudah dipakai (kolom UNIQUE)
		http.Error(w, "Username already taken or DB error", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Data updated successfully",
		"username": req.Username,
	})
}

func GetMyAuditLog(w http.ResponseWriter, r *http.Request) {
	user, _ := lmodel.UserFromContext(r.Context())

	logs, err := db.GetUserAuditLogs(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch audit logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":  user.Username,
		"total": len(logs),
		"logs":  logs,
	})
}

func ExportMyData(w http.ResponseWriter, r *http.Request) {
	user, _ := lmodel.UserFromContext(r.Context())

	profile, err := db.GetUserByID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	projects, err := db.GetUserProjects(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)
		return
	}

	logs, err := db.GetUserAuditLogs(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch audit logs", http.StatusInternalServerError)
		return
	}

	export := map[string]interface{}{
		"exported_at": time.Now(),
		"profile":     profile,
		"projects":    projects,
		"audit_logs":  logs,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=my_data_export.json")
	json.NewEncoder(w).Encode(export)
}
func DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	user, _ := lmodel.UserFromContext(r.Context())

	//  audit log milik user dihapus (di skema: ON DELETE SET NULL, jadi tidak ikut terhapus otomatis)
	if _, err := db.Pool.Exec(r.Context(), `DELETE FROM audit_logs WHERE user_id = $1`, user.ID); err != nil {
		http.Error(w, "Failed to delete audit logs", http.StatusInternalServerError)
		return
	}

	// projects & grades ikut terhapus otomatis (ON DELETE CASCADE)
	if err := db.DeleteUserAccount(r.Context(), user.ID); err != nil {
		http.Error(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Your account and all data have been permanently deleted",
	})
}
