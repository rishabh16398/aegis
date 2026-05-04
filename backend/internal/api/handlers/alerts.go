package handlers

import (
	"net/http"

	"github.com/aegis-av/aegis/internal/api/respond"
	"github.com/aegis-av/aegis/internal/db"
	"github.com/aegis-av/aegis/internal/db/models"
	"github.com/go-chi/chi/v5"
)

type AlertHandler struct {
	db *db.DB
}

func NewAlertHandler(database *db.DB) *AlertHandler {
	return &AlertHandler{db: database}
}

func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	var alerts []models.Alert
	h.db.Conn().Order("created_at desc").Limit(100).Find(&alerts)
	respond.JSON(w, http.StatusOK, alerts)
}

func (h *AlertHandler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result := h.db.Conn().Model(&models.Alert{}).
		Where("id = ?", id).
		Update("status", "acknowledged")
	if result.RowsAffected == 0 {
		respond.Error(w, http.StatusNotFound, "alert not found")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
}

func (h *AlertHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result := h.db.Conn().Model(&models.Alert{}).
		Where("id = ?", id).
		Update("status", "dismissed")
	if result.RowsAffected == 0 {
		respond.Error(w, http.StatusNotFound, "alert not found")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]string{"status": "dismissed"})
}
