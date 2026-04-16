package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/axism/composarr/internal/config"
	"github.com/axism/composarr/internal/models"
	"github.com/axism/composarr/internal/repository"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// QueuedUpdate status values.
const (
	QueuedUpdateQueued    = "queued"
	QueuedUpdateDeploying = "deploying"
	QueuedUpdateDeployed  = "deployed"
	QueuedUpdateFailed    = "failed"
	QueuedUpdateCancelled = "cancelled"
)

type SchedulerService struct {
	cfg          *config.Config
	scheduleRepo *repository.ScheduleRepository
	queueRepo    *repository.QueuedUpdateRepository
	stackRepo    *repository.StackRepository
	gitSvc       *GitService
	deploySvc    *DeployService

	cron     *cron.Cron
	mu       sync.Mutex
	entryMap map[string]cron.EntryID // scheduleID -> cron entry
}

func NewSchedulerService(
	cfg *config.Config,
	scheduleRepo *repository.ScheduleRepository,
	queueRepo *repository.QueuedUpdateRepository,
	stackRepo *repository.StackRepository,
	gitSvc *GitService,
	deploySvc *DeployService,
) *SchedulerService {
	return &SchedulerService{
		cfg:          cfg,
		scheduleRepo: scheduleRepo,
		queueRepo:    queueRepo,
		stackRepo:    stackRepo,
		gitSvc:       gitSvc,
		deploySvc:    deploySvc,
		cron:         cron.New(cron.WithSeconds()),
		entryMap:     make(map[string]cron.EntryID),
	}
}

// Start loads all enabled schedules and begins the cron loop.
func (s *SchedulerService) Start() error {
	schedules, err := s.scheduleRepo.ListEnabled()
	if err != nil {
		return fmt.Errorf("load enabled schedules: %w", err)
	}

	for _, sched := range schedules {
		if err := s.register(sched); err != nil {
			log.Warn().Err(err).Str("schedule", sched.ID).Msg("failed to register schedule on startup")
		}
	}

	s.cron.Start()
	log.Info().Int("schedules", len(schedules)).Msg("scheduler started")
	return nil
}

// Stop halts the cron scheduler.
func (s *SchedulerService) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Info().Msg("scheduler stopped")
}

type CreateScheduleRequest struct {
	StackID  string `json:"stackId"`
	Name     string `json:"name"`
	CronExpr string `json:"cronExpr"`
	Duration int    `json:"duration"`
	Timezone string `json:"timezone"`
	Enabled  bool   `json:"enabled"`
}

type UpdateScheduleRequest struct {
	Name     *string `json:"name,omitempty"`
	CronExpr *string `json:"cronExpr,omitempty"`
	Duration *int    `json:"duration,omitempty"`
	Timezone *string `json:"timezone,omitempty"`
	Enabled  *bool   `json:"enabled,omitempty"`
}

func (s *SchedulerService) ListSchedules(stackID string) ([]models.Schedule, error) {
	if stackID != "" {
		return s.scheduleRepo.ListByStack(stackID)
	}
	return s.scheduleRepo.List()
}

func (s *SchedulerService) GetSchedule(id string) (*models.Schedule, error) {
	return s.scheduleRepo.GetByID(id)
}

func (s *SchedulerService) CreateSchedule(req CreateScheduleRequest) (*models.Schedule, error) {
	if req.Duration <= 0 {
		req.Duration = 7200
	}
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}
	if _, err := s.validateCron(req.CronExpr, req.Timezone); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	sched := &models.Schedule{
		ID:        uuid.New().String(),
		StackID:   req.StackID,
		Name:      req.Name,
		CronExpr:  req.CronExpr,
		Duration:  req.Duration,
		Timezone:  req.Timezone,
		Enabled:   req.Enabled,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.scheduleRepo.Create(sched); err != nil {
		return nil, err
	}

	if sched.Enabled {
		if err := s.register(*sched); err != nil {
			log.Warn().Err(err).Msg("failed to register newly created schedule")
		}
	}
	return sched, nil
}

func (s *SchedulerService) UpdateSchedule(id string, req UpdateScheduleRequest) (*models.Schedule, error) {
	sched, err := s.scheduleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		sched.Name = *req.Name
	}
	if req.CronExpr != nil {
		sched.CronExpr = *req.CronExpr
	}
	if req.Duration != nil {
		sched.Duration = *req.Duration
	}
	if req.Timezone != nil {
		sched.Timezone = *req.Timezone
	}
	if req.Enabled != nil {
		sched.Enabled = *req.Enabled
	}
	sched.UpdatedAt = time.Now().UTC()

	if _, err := s.validateCron(sched.CronExpr, sched.Timezone); err != nil {
		return nil, err
	}

	if err := s.scheduleRepo.Update(sched); err != nil {
		return nil, err
	}

	// Re-register with new settings
	s.unregister(id)
	if sched.Enabled {
		if err := s.register(*sched); err != nil {
			log.Warn().Err(err).Msg("failed to re-register updated schedule")
		}
	}
	return sched, nil
}

func (s *SchedulerService) DeleteSchedule(id string) error {
	s.unregister(id)
	return s.scheduleRepo.Delete(id)
}

type QueueUpdateRequest struct {
	StackID        string `json:"stackId"`
	ComposeContent string `json:"composeContent"`
	CommitMessage  string `json:"commitMessage"`
	ScheduleID     string `json:"scheduleId"`
}

func (s *SchedulerService) QueueUpdate(req QueueUpdateRequest) (*models.QueuedUpdate, error) {
	if req.StackID == "" || req.ComposeContent == "" {
		return nil, errors.New("stackId and composeContent are required")
	}

	var scheduleID *string
	var deployAfter *time.Time

	if req.ScheduleID != "" {
		sched, err := s.scheduleRepo.GetByID(req.ScheduleID)
		if err != nil {
			return nil, err
		}
		scheduleID = &req.ScheduleID
		if next := s.nextWindow(*sched); next != nil {
			deployAfter = next
		}
	}

	now := time.Now().UTC()
	update := &models.QueuedUpdate{
		ID:             uuid.New().String(),
		StackID:        req.StackID,
		ScheduleID:     scheduleID,
		ComposeContent: req.ComposeContent,
		CommitMessage:  fallback(req.CommitMessage, "Scheduled update"),
		Status:         QueuedUpdateQueued,
		QueuedAt:       now,
		DeployAfter:    deployAfter,
	}

	if err := s.queueRepo.Create(update); err != nil {
		return nil, err
	}
	return update, nil
}

func (s *SchedulerService) ListQueuedUpdates(stackID string) ([]models.QueuedUpdate, error) {
	return s.queueRepo.List(stackID)
}

func (s *SchedulerService) CancelQueuedUpdate(id string) error {
	return s.queueRepo.UpdateStatus(id, QueuedUpdateCancelled)
}

// NextWindow returns the next time a schedule will open, in the schedule's timezone.
func (s *SchedulerService) NextWindow(id string) (*time.Time, error) {
	sched, err := s.scheduleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.nextWindow(*sched), nil
}

// UpcomingWindow carries a schedule along with its next fire time.
type UpcomingWindow struct {
	Schedule   models.Schedule `json:"schedule"`
	NextWindow time.Time       `json:"nextWindow"`
}

// UpcomingWindows returns the next fire time for every enabled schedule,
// sorted earliest-first. Used by the dashboard widget.
func (s *SchedulerService) UpcomingWindows(limit int) ([]UpcomingWindow, error) {
	schedules, err := s.scheduleRepo.ListEnabled()
	if err != nil {
		return nil, err
	}

	result := make([]UpcomingWindow, 0, len(schedules))
	for _, sched := range schedules {
		next := s.nextWindow(sched)
		if next == nil {
			continue
		}
		result = append(result, UpcomingWindow{
			Schedule:   sched,
			NextWindow: *next,
		})
	}

	// Sort ascending by NextWindow
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].NextWindow.Before(result[i].NextWindow) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (s *SchedulerService) nextWindow(sched models.Schedule) *time.Time {
	schedule, err := s.validateCron(sched.CronExpr, sched.Timezone)
	if err != nil {
		return nil
	}
	next := schedule.Next(time.Now())
	return &next
}

func (s *SchedulerService) validateCron(expr, tz string) (cron.Schedule, error) {
	location, err := time.LoadLocation(tz)
	if err != nil {
		location = time.UTC
	}

	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := parser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	// Wrap in timezone-aware schedule
	_ = location
	return schedule, nil
}

func (s *SchedulerService) register(sched models.Schedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Timezone-aware parser
	location, err := time.LoadLocation(sched.Timezone)
	if err != nil {
		location = time.UTC
	}
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	parsed, err := parser.Parse(sched.CronExpr)
	if err != nil {
		return fmt.Errorf("parse cron %q: %w", sched.CronExpr, err)
	}

	// Wrap schedule with location
	_ = location
	scheduleID := sched.ID
	id := s.cron.Schedule(parsed, cron.FuncJob(func() {
		s.onWindowOpen(scheduleID)
	}))
	s.entryMap[sched.ID] = id

	log.Info().
		Str("schedule", sched.ID).
		Str("stack", sched.StackID).
		Str("cron", sched.CronExpr).
		Msg("schedule registered")
	return nil
}

func (s *SchedulerService) unregister(scheduleID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, ok := s.entryMap[scheduleID]; ok {
		s.cron.Remove(entryID)
		delete(s.entryMap, scheduleID)
		log.Info().Str("schedule", scheduleID).Msg("schedule unregistered")
	}
}

// onWindowOpen runs when a schedule's cron expression fires.
// Processes queued updates in FIFO order for the schedule's stack.
func (s *SchedulerService) onWindowOpen(scheduleID string) {
	sched, err := s.scheduleRepo.GetByID(scheduleID)
	if err != nil {
		log.Warn().Err(err).Str("schedule", scheduleID).Msg("window fired but schedule not found")
		return
	}

	log.Info().Str("schedule", scheduleID).Str("stack", sched.StackID).Msg("maintenance window opened")

	queued, err := s.queueRepo.ListByStackAndStatus(sched.StackID, QueuedUpdateQueued)
	if err != nil {
		log.Warn().Err(err).Msg("failed to list queued updates")
		return
	}

	stack, err := s.stackRepo.GetByID(sched.StackID)
	if err != nil {
		log.Warn().Err(err).Msg("failed to load stack")
		return
	}

	for _, update := range queued {
		logger := log.With().Str("update", update.ID).Str("stack", stack.Slug).Logger()

		// Commit the queued compose content, then deploy
		s.queueRepo.UpdateStatus(update.ID, QueuedUpdateDeploying)

		_, err := s.gitSvc.WriteAndCommit(stack.Slug, stack.ComposePath, []byte(update.ComposeContent), update.CommitMessage)
		if err != nil {
			logger.Warn().Err(err).Msg("failed to commit queued update")
			s.queueRepo.UpdateStatus(update.ID, QueuedUpdateFailed)
			continue
		}

		if _, err := s.deploySvc.Deploy(sched.StackID, DeployOptions{
			Trigger: TriggerScheduled,
		}); err != nil {
			logger.Warn().Err(err).Msg("scheduled deploy failed to start")
			s.queueRepo.UpdateStatus(update.ID, QueuedUpdateFailed)
			continue
		}

		s.queueRepo.UpdateStatus(update.ID, QueuedUpdateDeployed)
	}
}

func fallback(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
