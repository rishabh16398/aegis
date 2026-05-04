package ws

// EventType identifies what kind of real-time update is being sent.
type EventType string

const (
	EventScanProgress  EventType = "scan.progress"
	EventScanComplete  EventType = "scan.complete"
	EventThreatFound   EventType = "threat.found"
	EventAlertNew      EventType = "alert.new"
	EventNetworkThreat EventType = "network.threat"
	EventProcessFlag   EventType = "process.flagged"
)

// Event is the JSON envelope sent over WebSocket to every connected client.
type Event struct {
	Type    EventType `json:"type"`
	Payload any       `json:"payload"`
}
