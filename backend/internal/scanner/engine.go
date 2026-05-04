package scanner

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aegis-av/aegis/internal/alert"
	"github.com/aegis-av/aegis/internal/api/ws"
	"github.com/aegis-av/aegis/internal/db"
	"github.com/aegis-av/aegis/internal/db/models"
	"github.com/google/uuid"
)

type Engine struct {
	workers        int
	quarantinePath string
	db             *db.DB
	hub            *ws.Hub
	alertEng       *alert.Engine
	clamav         *clamavClient
	clamavReady    bool
}

func New(workers int, quarantinePath string, database *db.DB, hub *ws.Hub, alertEng *alert.Engine) *Engine {
	client := newClamAVClient("")
	ready := true
	if err := client.Ping(); err != nil {
		slog.Warn("ClamAV daemon not reachable — virus scanning disabled", "err", err)
		slog.Warn("Install & start: brew install clamav && brew services start clamav")
		ready = false
	} else {
		slog.Info("ClamAV daemon connected")
	}

	return &Engine{
		workers:        workers,
		quarantinePath: quarantinePath,
		db:             database,
		hub:            hub,
		alertEng:       alertEng,
		clamav:         client,
		clamavReady:    ready,
	}
}

// StartScan creates a ScanJob in the DB and runs the scan in the background.
// Returns the job ID immediately so the handler can return it to the client.
func (e *Engine) StartScan(path string, recursive bool) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("path not found: %w", err)
	}

	jobType := "file"
	if info.IsDir() {
		if recursive {
			jobType = "directory"
		} else {
			jobType = "directory"
		}
	}

	now := time.Now()
	job := models.ScanJob{
		ID:        uuid.NewString(),
		Type:      jobType,
		Status:    "running",
		Path:      path,
		StartedAt: &now,
	}
	if err := e.db.Conn().Create(&job).Error; err != nil {
		return "", fmt.Errorf("create job: %w", err)
	}

	go e.runJob(job.ID, path, recursive)

	return job.ID, nil
}

// runJob collects files, scans them with the worker pool, then marks the job done.
func (e *Engine) runJob(jobID, path string, recursive bool) {
	files, err := collectFiles(path, recursive)
	if err != nil {
		slog.Error("scanner: collect files", "err", err)
		e.markJob(jobID, "failed", 0, 0)
		return
	}

	e.db.Conn().Model(&models.ScanJob{}).
		Where("id = ?", jobID).
		Update("total_files", len(files))

	slog.Info("scan started", "job", jobID, "files", len(files))

	// Feed files into a buffered channel — workers drain it.
	filesCh := make(chan string, len(files))
	for _, f := range files {
		filesCh <- f
	}
	close(filesCh)

	var (
		scanned atomic.Int64
		threats atomic.Int64
		wg      sync.WaitGroup
	)

	for i := 0; i < e.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range filesCh {
				isClean := e.processFile(jobID, filePath)
				scanned.Add(1)
				if !isClean {
					threats.Add(1)
				}

				// Periodic DB progress update (every 10 files).
				s := scanned.Load()
				if s%10 == 0 || s == int64(len(files)) {
					e.db.Conn().Model(&models.ScanJob{}).
						Where("id = ?", jobID).
						Updates(map[string]any{
							"scanned_files": s,
							"threats_found": threats.Load(),
						})
				}

				// Push live progress to frontend.
				e.hub.Publish(ws.Event{
					Type: ws.EventScanProgress,
					Payload: map[string]any{
						"job_id":  jobID,
						"scanned": s,
						"total":   len(files),
					},
				})
			}
		}()
	}

	wg.Wait()
	e.markJob(jobID, "completed", int(scanned.Load()), int(threats.Load()))
	slog.Info("scan complete", "job", jobID, "threats", threats.Load())

	e.hub.Publish(ws.Event{
		Type: ws.EventScanComplete,
		Payload: map[string]any{
			"job_id":        jobID,
			"files_scanned": scanned.Load(),
			"threats_found": threats.Load(),
		},
	})
}

// processFile scans one file, writes a ScanResult, and raises an alert if needed.
// Returns true if the file is clean.
func (e *Engine) processFile(jobID, filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		e.saveResult(jobID, filePath, "error", "", "", "", "", 0)
		return true
	}

	hashes, err := computeHashes(filePath)
	if err != nil {
		e.saveResult(jobID, filePath, "error", "", "", hashes.MD5, hashes.SHA256, info.Size())
		return true
	}

	// ClamAV scan.
	if e.clamavReady {
		result, err := e.clamav.ScanFile(filePath)
		if err != nil {
			slog.Debug("clamav scan error", "file", filePath, "err", err)
		} else if !result.Clean {
			e.saveResult(jobID, filePath, "threat", result.ThreatName, "clamav", hashes.MD5, hashes.SHA256, info.Size())
			e.raiseThreat(filePath, result.ThreatName, "clamav")
			return false
		}
	}

	e.saveResult(jobID, filePath, "clean", "", "", hashes.MD5, hashes.SHA256, info.Size())
	return true
}

func (e *Engine) saveResult(jobID, filePath, status, threatName, engine, md5, sha256 string, size int64) {
	result := models.ScanResult{
		JobID:      jobID,
		FilePath:   filePath,
		Status:     status,
		ThreatName: threatName,
		Engine:     engine,
		HashMD5:    md5,
		HashSHA256: sha256,
		FileSize:   size,
		ScannedAt:  time.Now(),
	}
	if err := e.db.Conn().Create(&result).Error; err != nil {
		slog.Error("save scan result", "err", err)
	}
}

func (e *Engine) raiseThreat(filePath, threatName, engine string) {
	threat := models.Threat{
		Name:       threatName,
		Type:       "malware",
		Severity:   classifySeverity(threatName),
		FilePath:   filePath,
		Source:     "file",
		Status:     "active",
		DetectedAt: time.Now(),
	}
	e.db.Conn().Create(&threat)

	e.hub.Publish(ws.Event{
		Type:    ws.EventThreatFound,
		Payload: threat,
	})

	e.alertEng.Raise(
		fmt.Sprintf("Threat detected: %s", threatName),
		fmt.Sprintf("File: %s", filePath),
		threat.Severity,
		"file",
		threat.ID,
	)
}

func (e *Engine) markJob(jobID, status string, scanned, threats int) {
	now := time.Now()
	e.db.Conn().Model(&models.ScanJob{}).
		Where("id = ?", jobID).
		Updates(map[string]any{
			"status":        status,
			"scanned_files": scanned,
			"threats_found": threats,
			"completed_at":  &now,
		})
}

// collectFiles returns all scannable files under path.
// If path is a file, returns just that file.
func collectFiles(path string, recursive bool) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}

	var files []string
	err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable dirs
		}
		if d.IsDir() {
			if !recursive && p != path {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type().IsRegular() {
			files = append(files, p)
		}
		return nil
	})
	return files, err
}

// classifySeverity makes a best-effort severity guess from the threat name.
func classifySeverity(name string) string {
	switch {
	case containsAny(name, "Ransomware", "Ransom", "Worm"):
		return "critical"
	case containsAny(name, "Trojan", "Backdoor", "Rootkit"):
		return "high"
	case containsAny(name, "Adware", "PUA", "Spyware"):
		return "medium"
	default:
		return "high"
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
