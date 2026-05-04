package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aegis-av/aegis/internal/api/respond"
	"github.com/aegis-av/aegis/internal/db"
	"github.com/aegis-av/aegis/internal/db/models"
	"github.com/aegis-av/aegis/internal/scanner"
	"github.com/go-chi/chi/v5"
)

type ScanHandler struct {
	engine *scanner.Engine
	db     *db.DB
}

func NewScanHandler(engine *scanner.Engine, database *db.DB) *ScanHandler {
	return &ScanHandler{engine: engine, db: database}
}

type scanRequest struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive"`
}

func (h *ScanHandler) StartFileScan(w http.ResponseWriter, r *http.Request) {
	var req scanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
		respond.Error(w, http.StatusBadRequest, "body must be {\"path\":\"...\"}")
		return
	}

	jobID, err := h.engine.StartScan(req.Path, false)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	respond.JSON(w, http.StatusAccepted, map[string]string{"job_id": jobID})
}

func (h *ScanHandler) StartDirScan(w http.ResponseWriter, r *http.Request) {
	var req scanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
		respond.Error(w, http.StatusBadRequest, "body must be {\"path\":\"...\",\"recursive\":true}")
		return
	}

	jobID, err := h.engine.StartScan(req.Path, req.Recursive)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	respond.JSON(w, http.StatusAccepted, map[string]string{"job_id": jobID})
}

func (h *ScanHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	var jobs []models.ScanJob
	h.db.Conn().Order("created_at desc").Limit(50).Find(&jobs)
	respond.JSON(w, http.StatusOK, jobs)
}

func (h *ScanHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var job models.ScanJob
	if err := h.db.Conn().First(&job, "id = ?", id).Error; err != nil {
		respond.Error(w, http.StatusNotFound, "job not found")
		return
	}
	respond.JSON(w, http.StatusOK, job)
}

func (h *ScanHandler) GetJobResults(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var results []models.ScanResult
	h.db.Conn().Where("job_id = ?", id).Order("scanned_at desc").Find(&results)
	respond.JSON(w, http.StatusOK, results)
}

func (h *ScanHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	respond.Error(w, http.StatusNotImplemented, "job cancellation not yet implemented")
}
