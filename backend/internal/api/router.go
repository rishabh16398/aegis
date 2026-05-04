package api

import (
	"net/http"

	"github.com/aegis-av/aegis/internal/api/handlers"
	"github.com/aegis-av/aegis/internal/api/respond"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	// ── Middleware stack ──────────────────────────────────────────────────────
	// Every request passes through these in order, top to bottom.

	// Recover from panics and return 500 instead of crashing the server.
	r.Use(middleware.Recoverer)

	// Log every request: method, path, status, duration.
	r.Use(middleware.Logger)

	// Real IP — populate r.RemoteAddr correctly behind proxies.
	r.Use(middleware.RealIP)

	// CORS — allow the React dev server (localhost:5173) to talk to us.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Authorization"},
	}))

	// ── Routes ───────────────────────────────────────────────────────────────

	scan       := handlers.NewScanHandler()
	network    := handlers.NewNetworkHandler()
	process    := handlers.NewProcessHandler()
	quarantine := handlers.NewQuarantineHandler()
	alert      := handlers.NewAlertHandler()
	stats      := handlers.NewStatsHandler()

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Get("/api/v1/stats", stats.Get)

	// Scan
	r.Route("/api/v1/scan", func(r chi.Router) {
		r.Post("/file", scan.StartFileScan)
		r.Post("/directory", scan.StartDirScan)
		r.Get("/jobs", scan.ListJobs)
		r.Get("/jobs/{id}", scan.GetJob)
		r.Get("/jobs/{id}/results", scan.GetJobResults)
		r.Delete("/jobs/{id}", scan.CancelJob)
	})

	// Network
	r.Route("/api/v1/network", func(r chi.Router) {
		r.Get("/connections", network.ListConnections)
		r.Get("/events", network.ListEvents)
		r.Post("/block", network.BlockIP)
		r.Get("/blocked", network.ListBlockedIPs)
		r.Delete("/blocked/{ip}", network.UnblockIP)
	})

	// Processes
	r.Route("/api/v1/processes", func(r chi.Router) {
		r.Get("/", process.ListProcesses)
		r.Get("/suspicious", process.ListSuspicious)
	})

	// Quarantine
	r.Route("/api/v1/quarantine", func(r chi.Router) {
		r.Get("/", quarantine.List)
		r.Post("/{id}/restore", quarantine.Restore)
		r.Delete("/{id}", quarantine.Delete)
	})

	// Alerts
	r.Route("/api/v1/alerts", func(r chi.Router) {
		r.Get("/", alert.List)
		r.Put("/{id}/ack", alert.Acknowledge)
		r.Delete("/{id}", alert.Dismiss)
	})

	return r
}
