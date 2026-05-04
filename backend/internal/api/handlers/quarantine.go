package handlers

import (
	"net/http"

	"github.com/aegis-av/aegis/internal/api/respond"
)

type QuarantineHandler struct{}

func NewQuarantineHandler() *QuarantineHandler { return &QuarantineHandler{} }

func (h *QuarantineHandler) List(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, []any{})
}

func (h *QuarantineHandler) Restore(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "quarantine manager not wired yet (Task 10)")
}

func (h *QuarantineHandler) Delete(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "quarantine manager not wired yet (Task 10)")
}
