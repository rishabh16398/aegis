package handlers

import (
	"net/http"

	"github.com/aegis-av/aegis/internal/api/respond"
)

type ScanHandler struct{}

func NewScanHandler() *ScanHandler { return &ScanHandler{} }

func (h *ScanHandler) StartFileScan(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "file scanner not wired yet (Task 5)")
}

func (h *ScanHandler) StartDirScan(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "file scanner not wired yet (Task 5)")
}

func (h *ScanHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, []any{})
}

func (h *ScanHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "not implemented")
}

func (h *ScanHandler) GetJobResults(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "not implemented")
}

func (h *ScanHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "not implemented")
}
