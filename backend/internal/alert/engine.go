package alert

import (
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/aegis-av/aegis/internal/api/ws"
	"github.com/aegis-av/aegis/internal/db"
	"github.com/aegis-av/aegis/internal/db/models"
)

type Engine struct {
	db  *db.DB
	hub *ws.Hub
}

func New(database *db.DB, hub *ws.Hub) *Engine {
	return &Engine{db: database, hub: hub}
}

// Raise creates an Alert in the DB, pushes it over WebSocket,
// and fires a macOS system notification.
func (e *Engine) Raise(title, message, severity, source, refID string) {
	alert := models.Alert{
		Title:    title,
		Message:  message,
		Severity: severity,
		Source:   source,
		RefID:    refID,
		Status:   "unread",
	}

	if err := e.db.Conn().Create(&alert).Error; err != nil {
		slog.Error("alert: failed to save", "err", err)
		return
	}

	e.hub.Publish(ws.Event{
		Type:    ws.EventAlertNew,
		Payload: alert,
	})

	e.notify(title, message, severity)
}

// notify sends a macOS Notification Center alert using osascript.
// It never blocks or crashes if the command fails.
func (e *Engine) notify(title, message, severity string) {
	sound := notificationSound(severity)
	script := fmt.Sprintf(
		`display notification %q with title "Aegis — %s" sound name %q`,
		message, title, sound,
	)
	if err := exec.Command("osascript", "-e", script).Run(); err != nil {
		slog.Debug("macos notification failed", "err", err)
	}
}

func notificationSound(severity string) string {
	switch severity {
	case "critical":
		return "Sosumi"
	case "high":
		return "Basso"
	default:
		return "Ping"
	}
}
