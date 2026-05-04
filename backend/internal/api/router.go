package api

import (
	"net/http"

	"github.com/aegis-av/aegis/internal/api/handlers"
	"github.com/aegis-av/aegis/internal/api/respond"
	"github.com/aegis-av/aegis/internal/api/ws"
	"github.com/aegis-av/aegis/internal/db"
	"github.com/aegis-av/aegis/internal/scanner"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(hub *ws.Hub, scanEngine *scanner.Engine, database *db.DB) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Authorization"},
	}))

	scan       := handlers.NewScanHandler(scanEngine, database)
	network    := handlers.NewNetworkHandler()
	process    := handlers.NewProcessHandler()
	quarantine := handlers.NewQuarantineHandler()
	alert      := handlers.NewAlertHandler(database)
	stats      := handlers.NewStatsHandler(database)

	r.Get("/ws", hub.ServeWS)

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Get("/api/v1/stats", stats.Get)

	r.Route("/api/v1/scan", func(r chi.Router) {
		r.Post("/file", scan.StartFileScan)
		r.Post("/directory", scan.StartDirScan)
		r.Get("/jobs", scan.ListJobs)
		r.Get("/jobs/{id}", scan.GetJob)
		r.Get("/jobs/{id}/results", scan.GetJobResults)
		r.Delete("/jobs/{id}", scan.CancelJob)
	})

	r.Route("/api/v1/network", func(r chi.Router) {
		r.Get("/connections", network.ListConnections)
		r.Get("/events", network.ListEvents)
		r.Post("/block", network.BlockIP)
		r.Get("/blocked", network.ListBlockedIPs)
		r.Delete("/blocked/{ip}", network.UnblockIP)
	})

	r.Route("/api/v1/processes", func(r chi.Router) {
		r.Get("/", process.ListProcesses)
		r.Get("/suspicious", process.ListSuspicious)
	})

	r.Route("/api/v1/quarantine", func(r chi.Router) {
		r.Get("/", quarantine.List)
		r.Post("/{id}/restore", quarantine.Restore)
		r.Delete("/{id}", quarantine.Delete)
	})

	r.Route("/api/v1/alerts", func(r chi.Router) {
		r.Get("/", alert.List)
		r.Put("/{id}/ack", alert.Acknowledge)
		r.Delete("/{id}", alert.Dismiss)
	})

	return r
}
