package handlers

import (
	"net/http"

	"github.com/aegis-av/aegis/internal/api/respond"
)

type AlertHandler struct{}

func NewAlertHandler() *AlertHandler { return &AlertHandler{} }

func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, []any{})
}

func (h *AlertHandler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "not implemented")
}

func (h *AlertHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "not implemented")
}
