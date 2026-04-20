package models

type Schedule struct {
	ID        string `db:"id" json:"id"`
	StackID   string `db:"stack_id" json:"stackId"`
	Name      string `db:"name" json:"name"`
	CronExpr  string `db:"cron_expr" json:"cronExpr"`
	Duration  int    `db:"duration" json:"duration"`
	Timezone  string `db:"timezone" json:"timezone"`
	Enabled   bool   `db:"enabled" json:"enabled"`
	CreatedAt Time   `db:"created_at" json:"createdAt"`
	UpdatedAt Time   `db:"updated_at" json:"updatedAt"`
}

type QueuedUpdate struct {
	ID             string  `db:"id" json:"id"`
	StackID        string  `db:"stack_id" json:"stackId"`
	ScheduleID     *string `db:"schedule_id" json:"scheduleId"`
	ComposeContent string  `db:"compose_content" json:"composeContent"`
	CommitMessage  string  `db:"commit_message" json:"commitMessage"`
	Status         string  `db:"status" json:"status"`
	QueuedAt       Time    `db:"queued_at" json:"queuedAt"`
	DeployAfter    *Time   `db:"deploy_after" json:"deployAfter"`
	DeployedAt     *Time   `db:"deployed_at" json:"deployedAt"`
}
