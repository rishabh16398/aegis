package handlers

import (
	"net/http"

	"github.com/aegis-av/aegis/internal/api/respond"
)

type ProcessHandler struct{}

func NewProcessHandler() *ProcessHandler { return &ProcessHandler{} }

func (h *ProcessHandler) ListProcesses(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, []any{})
}

func (h *ProcessHandler) ListSuspicious(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, []any{})
}
