// Package task provides task management functionality
package task

import (
	"aibuddy/pkg/config"
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
)

// Manager 任务管理器
type Manager struct {
	client    *asynq.Client
	scheduler *asynq.Scheduler
	server    *asynq.Server
	inspector *asynq.Inspector
	redisOpt  asynq.RedisClientOpt
}

// NewManager 创建任务管理器
func NewManager() *Manager {
	redisConfig := config.Instance.Storage.Redis
	if redisConfig == nil {
		panic("Task manager requires redis config")
	}

	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Username: redisConfig.Username,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	}

	client := asynq.NewClient(redisOpt)

	scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		Location: time.Local,
		LogLevel: asynq.ErrorLevel,
		EnqueueErrorHandler: func(_ *asynq.Task, _ []asynq.Option, err error) {
			log.Printf("Enqueue error: %v", err)
		},
	})

	inspector := asynq.NewInspector(redisOpt)

	return &Manager{
		client:    client,
		scheduler: scheduler,
		inspector: inspector,
		redisOpt:  redisOpt,
	}
}

// Client returns the asynq client for enqueuing tasks
func (m *Manager) Client() *asynq.Client {
	return m.client
}

// RedisOpt returns the redis options for creating a server
func (m *Manager) RedisOpt() asynq.RedisClientOpt {
	return m.redisOpt
}

// Scheduler returns the asynq scheduler for periodic tasks
func (m *Manager) Scheduler() *asynq.Scheduler {
	return m.scheduler
}

// Enqueue enqueues a task immediately
func (m *Manager) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return m.client.Enqueue(task, opts...)
}

// EnqueueIn enqueues a task to be processed after the specified duration
func (m *Manager) EnqueueIn(task *asynq.Task, d time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	newOpts := append([]asynq.Option{asynq.ProcessIn(d)}, opts...)
	return m.client.Enqueue(task, newOpts...)
}

// EnqueueAt enqueues a task to be processed at the specified time
func (m *Manager) EnqueueAt(task *asynq.Task, t time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	newOpts := append([]asynq.Option{asynq.ProcessAt(t)}, opts...)
	return m.client.Enqueue(task, newOpts...)
}

// RegisterPeriodicTask registers a periodic task with cron spec
// Returns entryID for later unregistration
func (m *Manager) RegisterPeriodicTask(cronSpec string, task *asynq.Task, opts ...asynq.Option) (string, error) {
	entryID, err := m.scheduler.Register(cronSpec, task, opts...)
	return entryID, err
}

// UnregisterPeriodicTask cancels a periodic task by entryID
func (m *Manager) UnregisterPeriodicTask(entryID string) error {
	return m.scheduler.Unregister(entryID)
}

// CancelTask deletes a pending task from the queue
func (m *Manager) CancelTask(queue, taskID string) error {
	return m.inspector.DeleteTask(queue, taskID)
}

// GetTaskInfoByID get task info by taskId
func (m *Manager) GetTaskInfoByID(query, taskID string) (*asynq.TaskInfo, error) {
	return m.inspector.GetTaskInfo(query, taskID)
}

// ClearQueue clears all tasks from a queue
func (m *Manager) ClearQueue(queue string) error {
	if m.inspector == nil {
		return nil
	}
	// Delete all pending tasks
	pendingTasks, err := m.inspector.ListPendingTasks(queue)
	if err != nil {
		return err
	}
	for _, task := range pendingTasks {
		_ = m.inspector.DeleteTask(queue, task.ID)
	}
	// Delete all scheduled tasks
	scheduledTasks, err := m.inspector.ListScheduledTasks(queue)
	if err != nil {
		return err
	}
	for _, task := range scheduledTasks {
		_ = m.inspector.DeleteTask(queue, task.ID)
	}
	return nil
}

// Start starts the scheduler
func (m *Manager) Start() error {
	return m.scheduler.Run()
}

// Shutdown gracefully closes the client, scheduler and server
func (m *Manager) Shutdown() error {
	m.scheduler.Shutdown()
	if m.server != nil {
		m.server.Shutdown()
	}
	if m.inspector != nil {
		_ = m.inspector.Close()
	}
	return m.client.Close()
}

// HandlerFunc is the function signature for task handlers
type HandlerFunc func(context.Context, *asynq.Task) error

var handlers = make(map[string]HandlerFunc)

// RegisterHandler registers a task handler for a specific task type
func RegisterHandler(taskType string, handler HandlerFunc) {
	handlers[taskType] = handler
}

// Instance is the global task manager instance
var Instance *Manager

// StartTaskServer starts the task server
func StartTaskServer(ctx context.Context) error {
	redisConfig := config.Instance.Storage.Redis
	if redisConfig == nil {
		return fmt.Errorf("task server requires redis config")
	}

	tm := NewManager()
	Instance = tm

	srv := asynq.NewServer(
		tm.redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default":  6,
				"critical": 10,
				"low":      1,
			},
		},
	)
	tm.server = srv

	InitHandlers()

	mux := asynq.NewServeMux()
	for taskType, handler := range handlers {
		mux.HandleFunc(taskType, handler)
	}

	go func() {
		<-ctx.Done()
		slog.Info("Task server shutting down...")
		if err := tm.Shutdown(); err != nil {
			slog.Error("Failed to shutdown task server", "error", err)
		}
	}()

	go func() {
		if err := tm.scheduler.Run(); err != nil {
			slog.Error("Scheduler error", "error", err)
		}
	}()

	return srv.Run(mux)
}

// InitHandlers 初始化任务处理器
func InitHandlers() {
	// 时间提醒
	if _, exists := handlers[TaskTypeReminder]; !exists {
		RegisterHandler(TaskTypeReminder, ReminderHandler)
	}

	// 纪念日提醒
	if _, exists := handlers[TaskTypeAnniversary]; !exists {
		RegisterHandler(TaskTypeAnniversary, AnniversaryHandler)
	}
}
