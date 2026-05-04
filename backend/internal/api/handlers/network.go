package handlers

import (
	"net/http"

	"github.com/aegis-av/aegis/internal/api/respond"
)

type NetworkHandler struct{}

func NewNetworkHandler() *NetworkHandler { return &NetworkHandler{} }

func (h *NetworkHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, []any{})
}

func (h *NetworkHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, []any{})
}

func (h *NetworkHandler) BlockIP(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "network monitor not wired yet (Task 7)")
}

func (h *NetworkHandler) ListBlockedIPs(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, []any{})
}

func (h *NetworkHandler) UnblockIP(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "not implemented")
}
