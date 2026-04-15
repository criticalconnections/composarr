package models

import "time"

type Deployment struct {
	ID             string     `db:"id" json:"id"`
	StackID        string     `db:"stack_id" json:"stackId"`
	CommitHash     string     `db:"commit_hash" json:"commitHash"`
	PreviousCommit string     `db:"previous_commit" json:"previousCommit"`
	Status         string     `db:"status" json:"status"`
	TriggerType    string     `db:"trigger_type" json:"triggerType"`
	StartedAt      *time.Time `db:"started_at" json:"startedAt"`
	CompletedAt    *time.Time `db:"completed_at" json:"completedAt"`
	ErrorMessage   string     `db:"error_message" json:"errorMessage"`
	CreatedAt      time.Time  `db:"created_at" json:"createdAt"`
}

type DeploymentLog struct {
	ID           string    `db:"id" json:"id"`
	DeploymentID string    `db:"deployment_id" json:"deploymentId"`
	Level        string    `db:"level" json:"level"`
	Message      string    `db:"message" json:"message"`
	Timestamp    time.Time `db:"timestamp" json:"timestamp"`
}

type HealthCheckResult struct {
	ID            string    `db:"id" json:"id"`
	DeploymentID  string    `db:"deployment_id" json:"deploymentId"`
	ContainerName string    `db:"container_name" json:"containerName"`
	ServiceName   string    `db:"service_name" json:"serviceName"`
	Status        string    `db:"status" json:"status"`
	CheckOutput   string    `db:"check_output" json:"checkOutput"`
	CheckedAt     time.Time `db:"checked_at" json:"checkedAt"`
}
