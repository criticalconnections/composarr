package websocket

import "time"

// Event types broadcast to WebSocket clients.
const (
	EventDeployStarted         = "deploy.started"
	EventDeployValidating      = "deploy.validating"
	EventDeployPulling         = "deploy.pulling"
	EventDeployStarting        = "deploy.starting"
	EventDeployHealthChecking  = "deploy.health_checking"
	EventDeploySucceeded       = "deploy.succeeded"
	EventDeployFailed          = "deploy.failed"
	EventDeployRollingBack     = "deploy.rolling_back"
	EventDeployRolledBack      = "deploy.rolled_back"
	EventDeployLog             = "deploy.log"
	EventHealthUpdate          = "health.update"
	EventStackStatusChanged    = "stack.status_changed"
)

// Event represents a single push from the server to subscribed clients.
type Event struct {
	Type         string      `json:"type"`
	StackID      string      `json:"stackId,omitempty"`
	DeploymentID string      `json:"deploymentId,omitempty"`
	Payload      interface{} `json:"payload,omitempty"`
	Timestamp    time.Time   `json:"timestamp"`
}

// NewEvent constructs a new event with the current timestamp.
func NewEvent(eventType, stackID, deploymentID string, payload interface{}) Event {
	return Event{
		Type:         eventType,
		StackID:      stackID,
		DeploymentID: deploymentID,
		Payload:      payload,
		Timestamp:    time.Now().UTC(),
	}
}
