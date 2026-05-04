package handlers

import (
	"net/http"

	"github.com/aegis-av/aegis/internal/api/respond"
	"github.com/aegis-av/aegis/internal/db"
	"github.com/aegis-av/aegis/internal/db/models"
)

type StatsHandler struct {
	db *db.DB
}

func NewStatsHandler(database *db.DB) *StatsHandler {
	return &StatsHandler{db: database}
}

func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	var totalScans, threatsFound, filesScanned, alertsUnread, quarantined, blockedIPs int64

	h.db.Conn().Model(&models.ScanJob{}).Count(&totalScans)
	h.db.Conn().Model(&models.Threat{}).Count(&threatsFound)
	h.db.Conn().Model(&models.ScanResult{}).Where("status = ?", "threat").Count(&filesScanned)
	h.db.Conn().Model(&models.Alert{}).Where("status = ?", "unread").Count(&alertsUnread)
	h.db.Conn().Model(&models.QuarantineEntry{}).Where("restored_at IS NULL").Count(&quarantined)
	h.db.Conn().Model(&models.BlockedIP{}).Count(&blockedIPs)

	respond.JSON(w, http.StatusOK, map[string]int64{
		"total_scans":   totalScans,
		"threats_found": threatsFound,
		"files_scanned": filesScanned,
		"alerts_unread": alertsUnread,
		"quarantined":   quarantined,
		"blocked_ips":   blockedIPs,
	})
}
