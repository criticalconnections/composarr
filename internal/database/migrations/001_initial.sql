CREATE TABLE IF NOT EXISTS stacks (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,
    slug          TEXT NOT NULL UNIQUE,
    description   TEXT NOT NULL DEFAULT '',
    compose_path  TEXT NOT NULL DEFAULT 'docker-compose.yml',
    status        TEXT NOT NULL DEFAULT 'unknown',
    auto_update   INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS stack_dependencies (
    id              TEXT PRIMARY KEY,
    stack_id        TEXT NOT NULL REFERENCES stacks(id) ON DELETE CASCADE,
    depends_on_id   TEXT NOT NULL REFERENCES stacks(id) ON DELETE CASCADE,
    dependency_type TEXT NOT NULL DEFAULT 'hard',
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(stack_id, depends_on_id)
);

CREATE TABLE IF NOT EXISTS deployments (
    id              TEXT PRIMARY KEY,
    stack_id        TEXT NOT NULL REFERENCES stacks(id) ON DELETE CASCADE,
    commit_hash     TEXT NOT NULL DEFAULT '',
    previous_commit TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'pending',
    trigger_type    TEXT NOT NULL DEFAULT 'manual',
    started_at      TEXT,
    completed_at    TEXT,
    error_message   TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_deployments_stack_id ON deployments(stack_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);

CREATE TABLE IF NOT EXISTS deployment_logs (
    id            TEXT PRIMARY KEY,
    deployment_id TEXT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    level         TEXT NOT NULL DEFAULT 'info',
    message       TEXT NOT NULL,
    timestamp     TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_deployment_logs_deployment_id ON deployment_logs(deployment_id);

CREATE TABLE IF NOT EXISTS health_check_results (
    id              TEXT PRIMARY KEY,
    deployment_id   TEXT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    container_name  TEXT NOT NULL,
    service_name    TEXT NOT NULL,
    status          TEXT NOT NULL,
    check_output    TEXT NOT NULL DEFAULT '',
    checked_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_hcr_deployment_id ON health_check_results(deployment_id);

CREATE TABLE IF NOT EXISTS schedules (
    id          TEXT PRIMARY KEY,
    stack_id    TEXT NOT NULL REFERENCES stacks(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    cron_expr   TEXT NOT NULL,
    duration    INTEGER NOT NULL DEFAULT 7200,
    timezone    TEXT NOT NULL DEFAULT 'UTC',
    enabled     INTEGER NOT NULL DEFAULT 1,
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_schedules_stack_id ON schedules(stack_id);

CREATE TABLE IF NOT EXISTS queued_updates (
    id              TEXT PRIMARY KEY,
    stack_id        TEXT NOT NULL REFERENCES stacks(id) ON DELETE CASCADE,
    schedule_id     TEXT REFERENCES schedules(id) ON DELETE SET NULL,
    compose_content TEXT NOT NULL,
    commit_message  TEXT NOT NULL DEFAULT 'Scheduled update',
    status          TEXT NOT NULL DEFAULT 'queued',
    queued_at       TEXT NOT NULL DEFAULT (datetime('now')),
    deploy_after    TEXT,
    deployed_at     TEXT
);

CREATE INDEX IF NOT EXISTS idx_queued_updates_status ON queued_updates(status);
CREATE INDEX IF NOT EXISTS idx_queued_updates_stack_id ON queued_updates(stack_id);

CREATE TABLE IF NOT EXISTS stack_environment (
    id        TEXT PRIMARY KEY,
    stack_id  TEXT NOT NULL REFERENCES stacks(id) ON DELETE CASCADE,
    key       TEXT NOT NULL,
    value     TEXT NOT NULL,
    is_secret INTEGER NOT NULL DEFAULT 0,
    UNIQUE(stack_id, key)
);

CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);
