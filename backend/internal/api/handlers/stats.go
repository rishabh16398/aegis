package handlers

import (
	"net/http"

	"github.com/aegis-av/aegis/internal/api/respond"
)

type StatsHandler struct{}

func NewStatsHandler() *StatsHandler { return &StatsHandler{} }

func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, map[string]int{
		"total_scans":    0,
		"threats_found":  0,
		"files_scanned":  0,
		"alerts_unread":  0,
		"quarantined":    0,
		"blocked_ips":    0,
	})
}
