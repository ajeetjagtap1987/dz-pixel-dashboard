package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB

func main() {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env("PG_USER", "pixel_app"),
		os.Getenv("PG_PASSWORD"),
		env("PG_HOST", "localhost"),
		env("PG_PORT", "5432"),
		env("PG_DB", "argus_admin"),
	)

	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Printf("WARN pg open: %v", err)
	} else {
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			log.Printf("WARN pg ping: %v", err)
		} else {
			log.Println("postgres: connected")
			if err := ensureSchema(); err != nil {
				log.Printf("WARN schema: %v", err)
			}
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health)
	mux.HandleFunc("/api/stats", stats)
	mux.HandleFunc("/api/campaigns", campaigns)
	mux.HandleFunc("/api/whoami", whoami)
	mux.HandleFunc("/", root)

	addr := ":" + env("PORT", "8080")
	log.Printf("dashboard-svc listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}

func health(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	dbStatus := "down"
	if db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err == nil {
			dbStatus = "up"
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  status,
		"service": "dashboard-svc",
		"db":      dbStatus,
	})
}

func stats(w http.ResponseWriter, r *http.Request) {
	// Simple placeholder stats — replace with real aggregation queries
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"events_today":     12_543,
		"unique_users":     3_122,
		"active_campaigns": 7,
		"updated_at":       time.Now().Unix(),
		"note":             "placeholder — wire to ClickHouse for real numbers",
	})
}

type campaign struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func campaigns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if db == nil {
		_ = json.NewEncoder(w).Encode([]campaign{})
		return
	}

	rows, err := db.QueryContext(r.Context(), `SELECT id, name, status FROM campaigns ORDER BY id`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []campaign{}
	for rows.Next() {
		var c campaign
		if err := rows.Scan(&c.ID, &c.Name, &c.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out = append(out, c)
	}
	_ = json.NewEncoder(w).Encode(out)
}

func whoami(w http.ResponseWriter, r *http.Request) {
	// Placeholder — wire JWT auth here in real implementation
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"user":  "demo-admin",
		"role":  "admin",
		"email": "admin@example.com",
	})
}

func root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintln(w, "pixel-dashboard ready — API at /api/*")
}

func ensureSchema() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS campaigns (
			id         SERIAL PRIMARY KEY,
			name       TEXT NOT NULL,
			status     TEXT NOT NULL DEFAULT 'draft',
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		INSERT INTO campaigns (name, status)
			SELECT 'Welcome series', 'active'
			WHERE NOT EXISTS (SELECT 1 FROM campaigns WHERE name = 'Welcome series');
	`)
	return err
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", env("CORS_ORIGIN", "*"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
