package main

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var (
	db     *sql.DB
	apiKey string
)

func main() {
	apiKey = os.Getenv("CONTEXT_SHARE_API_KEY")
	if apiKey == "" {
		log.Fatal("CONTEXT_SHARE_API_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8787"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "context-share.db"
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	db.Exec("PRAGMA journal_mode=WAL")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS contexts (
		key TEXT PRIMARY KEY,
		context TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		expires_at INTEGER
	)`)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /context/{key}", withAuth(saveContext))
	mux.HandleFunc("GET /context/{key}", withAuth(loadContext))
	mux.HandleFunc("DELETE /context/{key}", withAuth(deleteContext))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("context-share server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if subtle.ConstantTimeCompare([]byte(token), []byte(apiKey)) != 1 {
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

type saveRequest struct {
	Context  json.RawMessage `json:"context"`
	TTLHours *int            `json:"ttl_hours,omitempty"`
}

func saveContext(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req saveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(req.Context) == 0 {
		jsonError(w, "context is required", http.StatusBadRequest)
		return
	}

	now := time.Now().Unix()
	var expiresAt *int64
	if req.TTLHours != nil {
		exp := time.Now().Add(time.Duration(*req.TTLHours) * time.Hour).Unix()
		expiresAt = &exp
	}

	_, err := db.Exec(
		`INSERT INTO contexts (key, context, created_at, expires_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET context=excluded.context, created_at=excluded.created_at, expires_at=excluded.expires_at`,
		key, string(req.Context), now, expiresAt,
	)
	if err != nil {
		jsonError(w, "failed to save", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"status": "saved", "key": key})
}

func loadContext(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var (
		ctxStr    string
		createdAt int64
		expiresAt sql.NullInt64
	)
	err := db.QueryRow(
		`SELECT context, created_at, expires_at FROM contexts WHERE key = ?`, key,
	).Scan(&ctxStr, &createdAt, &expiresAt)
	if err == sql.ErrNoRows {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, "failed to load", http.StatusInternalServerError)
		return
	}

	if expiresAt.Valid && expiresAt.Int64 < time.Now().Unix() {
		db.Exec(`DELETE FROM contexts WHERE key = ?`, key)
		jsonError(w, "context expired", http.StatusNotFound)
		return
	}

	resp := map[string]any{
		"key":        key,
		"context":    json.RawMessage(ctxStr),
		"created_at": time.Unix(createdAt, 0).UTC().Format(time.RFC3339),
	}
	if expiresAt.Valid {
		resp["expires_at"] = time.Unix(expiresAt.Int64, 0).UTC().Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func deleteContext(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	result, err := db.Exec(`DELETE FROM contexts WHERE key = ?`, key)
	if err != nil {
		jsonError(w, "failed to delete", http.StatusInternalServerError)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		jsonError(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "key": key})
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
