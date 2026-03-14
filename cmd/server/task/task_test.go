package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"aibuddy/pkg/config"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Task types
const (
	TaskTypeSendEmail     = "send_email"
	TaskTypeDelayedTask   = "delayed_task"
	TaskTypeScheduledTask = "scheduled_task"
	TaskTypeDailyReport   = "daily_report"
)

// Task payloads
type EmailPayload struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
}

type DelayedTaskPayload struct {
	Task string `json:"task"`
}

// Test fixtures
var (
	executedTasks sync.Map
	taskCounter   int64
)

func setupTest(t *testing.T) (*Manager, *asynq.Server, *asynq.Inspector) {
	// 加载配置
	config.Setup("")

	tm := NewManager()
	require.NotNil(t, tm)

	// 创建 Server
	srv := asynq.NewServer(
		tm.RedisOpt(),
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default":  6,
				"critical": 10,
				"low":      1,
			},
		},
	)

	// 创建 Inspector
	inspector := asynq.NewInspector(tm.RedisOpt())

	// 注册 handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskTypeSendEmail, handleSendEmail)
	mux.HandleFunc(TaskTypeDelayedTask, handleDelayedTask)
	mux.HandleFunc(TaskTypeScheduledTask, handleScheduledTask)
	mux.HandleFunc(TaskTypeDailyReport, handleDailyReport)

	// 启动 server
	go func() {
		srv.Run(mux)
	}()

	// 等待 server 完全启动
	time.Sleep(300 * time.Millisecond)

	return tm, srv, inspector
}

// Handler implementations
func handleSendEmail(ctx context.Context, t *asynq.Task) error {
	var payload EmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	taskKey := TaskTypeSendEmail + ":" + payload.Email
	executedTasks.Store(taskKey, payload)
	atomic.AddInt64(&taskCounter, 1)

	return nil
}

func handleDelayedTask(ctx context.Context, t *asynq.Task) error {
	var payload DelayedTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	taskKey := TaskTypeDelayedTask + ":" + payload.Task
	executedTasks.Store(taskKey, payload)
	atomic.AddInt64(&taskCounter, 1)

	return nil
}

func handleScheduledTask(ctx context.Context, t *asynq.Task) error {
	taskKey := TaskTypeScheduledTask + ":executed"
	executedTasks.Store(taskKey, true)
	atomic.AddInt64(&taskCounter, 1)
	return nil
}

func handleDailyReport(ctx context.Context, t *asynq.Task) error {
	taskKey := TaskTypeDailyReport + ":executed"
	executedTasks.Store(taskKey, true)
	atomic.AddInt64(&taskCounter, 1)
	return nil
}

// resetTestState clears test state
func resetTestState() {
	executedTasks = sync.Map{}
	atomic.StoreInt64(&taskCounter, 0)
	// 清空 handlers，确保测试隔离
	handlers = make(map[string]HandlerFunc)
}

// uniqueTaskID generates a unique task ID for testing
func uniqueTaskID(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), time.Now().Nanosecond())
}

func cleanup(tm *Manager, srv *asynq.Server, inspector *asynq.Inspector) {
	srv.Shutdown()
	tm.Shutdown()
	inspector.Close()
}

// Test cases

func TestManager_Enqueue(t *testing.T) {
	resetTestState()
	tm, srv, inspector := setupTest(t)
	defer cleanup(tm, srv, inspector)

	// 即时任务
	taskID := uniqueTaskID("email")
	payload, _ := json.Marshal(EmailPayload{
		Email:   "enqueue@test.com",
		Subject: "Hello",
	})
	task := asynq.NewTask(TaskTypeSendEmail, payload)

	info, err := tm.Enqueue(task, asynq.TaskID(taskID))
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, TaskTypeSendEmail, info.Type)
	assert.Equal(t, taskID, info.ID)

	// 等待任务执行
	assert.Eventually(t, func() bool {
		return atomic.LoadInt64(&taskCounter) >= 1
	}, 5*time.Second, 100*time.Millisecond, "任务应该被执行")

	// 验证任务被执行
	taskKey := TaskTypeSendEmail + ":enqueue@test.com"
	_, executed := executedTasks.Load(taskKey)
	assert.True(t, executed, "任务应该被正确处理")
}

func TestManager_EnqueueIn(t *testing.T) {
	resetTestState()
	tm, srv, inspector := setupTest(t)
	defer cleanup(tm, srv, inspector)

	// 延迟任务 - 入队到 scheduled set
	taskID := uniqueTaskID("delayed")
	payload, _ := json.Marshal(DelayedTaskPayload{Task: "delayed_job_test"})
	task := asynq.NewTask(TaskTypeDelayedTask, payload)

	info, err := tm.EnqueueIn(task, 10*time.Second, asynq.TaskID(taskID))
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, TaskTypeDelayedTask, info.Type)
	assert.Equal(t, taskID, info.ID)

	// 验证任务在 scheduled 集合中
	scheduledTasks, err := inspector.ListScheduledTasks("default", asynq.PageSize(100))
	require.NoError(t, err)

	var found bool
	for _, st := range scheduledTasks {
		if st.ID == taskID {
			found = true
			break
		}
	}
	assert.True(t, found, "延迟任务应该在 scheduled 集合中")

	// 删除 scheduled 任务
	err = inspector.DeleteTask("default", taskID)
	require.NoError(t, err)

	// 立即执行相同任务验证 handler 工作正常
	taskID2 := uniqueTaskID("delayed-immediate")
	task2 := asynq.NewTask(TaskTypeDelayedTask, payload)
	_, err = tm.Enqueue(task2, asynq.TaskID(taskID2))
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		return atomic.LoadInt64(&taskCounter) >= 1
	}, 5*time.Second, 100*time.Millisecond, "任务应该被执行")

	taskKey := TaskTypeDelayedTask + ":delayed_job_test"
	_, executed := executedTasks.Load(taskKey)
	assert.True(t, executed, "任务应该被正确处理")
}

func TestManager_EnqueueAt(t *testing.T) {
	resetTestState()
	tm, srv, inspector := setupTest(t)
	defer cleanup(tm, srv, inspector)

	// 定时任务
	taskID := uniqueTaskID("scheduled")
	processAt := time.Now().Add(10 * time.Second)
	payload := []byte(`{"scheduled": true}`)
	task := asynq.NewTask(TaskTypeScheduledTask, payload)

	info, err := tm.EnqueueAt(task, processAt, asynq.TaskID(taskID))
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, TaskTypeScheduledTask, info.Type)
	assert.Equal(t, taskID, info.ID)

	// 验证任务在 scheduled 集合中
	scheduledTasks, err := inspector.ListScheduledTasks("default", asynq.PageSize(100))
	require.NoError(t, err)

	var found bool
	for _, st := range scheduledTasks {
		if st.ID == taskID {
			found = true
			break
		}
	}
	assert.True(t, found, "定时任务应该在 scheduled 集合中")

	// 删除并立即执行验证 handler
	err = inspector.DeleteTask("default", taskID)
	require.NoError(t, err)

	taskID2 := uniqueTaskID("scheduled-immediate")
	task2 := asynq.NewTask(TaskTypeScheduledTask, payload)
	_, err = tm.Enqueue(task2, asynq.TaskID(taskID2))
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		return atomic.LoadInt64(&taskCounter) >= 1
	}, 5*time.Second, 100*time.Millisecond, "任务应该被执行")

	_, executed := executedTasks.Load(TaskTypeScheduledTask + ":executed")
	assert.True(t, executed, "任务应该被正确处理")
}

func TestManager_UniqueTask(t *testing.T) {
	resetTestState()
	tm, srv, inspector := setupTest(t)
	defer cleanup(tm, srv, inspector)

	// 测试任务唯一性
	email := fmt.Sprintf("unique-%d@test.com", time.Now().UnixNano())
	payload, _ := json.Marshal(EmailPayload{
		Email:   email,
		Subject: "Unique Test",
	})
	task := asynq.NewTask(TaskTypeSendEmail, payload)

	// 第一次入队应该成功
	info1, err := tm.Enqueue(task, asynq.Unique(time.Hour))
	require.NoError(t, err)
	require.NotNil(t, info1)

	// 第二次入队相同任务应该失败
	info2, err := tm.Enqueue(task, asynq.Unique(time.Hour))
	assert.Error(t, err, "相同任务应该因为唯一性约束而失败")
	assert.Nil(t, info2)
	assert.Contains(t, err.Error(), "already exists")

	// 等待第一次任务执行
	assert.Eventually(t, func() bool {
		return atomic.LoadInt64(&taskCounter) >= 1
	}, 5*time.Second, 100*time.Millisecond, "任务应该被执行一次")
}

func TestManager_TaskID_EnsureUniqueness(t *testing.T) {
	resetTestState()
	tm, srv, inspector := setupTest(t)
	defer cleanup(tm, srv, inspector)

	taskID := uniqueTaskID("same-id")

	// 第一个任务
	payload1, _ := json.Marshal(EmailPayload{
		Email:   "user1-taskid@test.com",
		Subject: "Test 1",
	})
	task1 := asynq.NewTask(TaskTypeSendEmail, payload1)

	info1, err := tm.Enqueue(task1, asynq.TaskID(taskID))
	require.NoError(t, err)
	assert.Equal(t, taskID, info1.ID)

	// 第二个任务使用相同 TaskID
	payload2, _ := json.Marshal(EmailPayload{
		Email:   "user2-taskid@test.com",
		Subject: "Test 2",
	})
	task2 := asynq.NewTask(TaskTypeSendEmail, payload2)

	info2, err := tm.Enqueue(task2, asynq.TaskID(taskID))
	assert.Error(t, err, "相同 TaskID 应该失败")
	assert.Nil(t, info2)

	// 等待任务执行
	assert.Eventually(t, func() bool {
		return atomic.LoadInt64(&taskCounter) >= 1
	}, 5*time.Second, 100*time.Millisecond)
}

func TestManager_RegisterPeriodicTask(t *testing.T) {
	resetTestState()
	tm, srv, inspector := setupTest(t)
	defer cleanup(tm, srv, inspector)

	taskID := uniqueTaskID("periodic")
	task := asynq.NewTask(TaskTypeDailyReport, nil)

	entryID, err := tm.RegisterPeriodicTask("* * * * *", task, asynq.TaskID(taskID))
	require.NoError(t, err)
	require.NotEmpty(t, entryID)

	// 等待任务执行
	assert.Eventually(t, func() bool {
		return atomic.LoadInt64(&taskCounter) >= 1
	}, 5*time.Second, 100*time.Millisecond, "任务应该被执行")

	// 取消任务
	err = tm.UnregisterPeriodicTask(entryID)
	require.NoError(t, err)

	// 验证任务没有被执行
	assert.Eventually(t, func() bool {
		return atomic.LoadInt64(&taskCounter) == 0
	}, 5*time.Second, 100*time.Millisecond, "任务应该没有被执行")
}

func TestManager_Client(t *testing.T) {
	resetTestState()
	tm, srv, inspector := setupTest(t)
	defer cleanup(tm, srv, inspector)

	client := tm.Client()
	assert.NotNil(t, client)
}

func TestManager_Scheduler(t *testing.T) {
	resetTestState()
	tm, srv, inspector := setupTest(t)
	defer cleanup(tm, srv, inspector)

	scheduler := tm.Scheduler()
	assert.NotNil(t, scheduler)
}

func TestManager_RedisOpt(t *testing.T) {
	resetTestState()
	tm, srv, inspector := setupTest(t)
	defer cleanup(tm, srv, inspector)

	redisOpt := tm.RedisOpt()
	assert.NotNil(t, redisOpt)
	assert.NotEmpty(t, redisOpt.Addr)
}

func TestStartTaskServer_Reminder(t *testing.T) {
	resetTestState()

	// 加载配置
	config.Setup("")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 注册测试用的 reminder handler
	var reminderExecuted atomic.Bool
	testReminderHandler := func(_ context.Context, task *asynq.Task) error {
		var payload ReminderTaskPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return err
		}
		slog.Info("[Reminder] executed with ", "remind_id", payload.RemindID)
		reminderExecuted.Store(true)
		atomic.AddInt64(&taskCounter, 1)
		return nil
	}
	handlers[TaskTypeReminder] = testReminderHandler

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- StartTaskServer(ctx)
	}()

	// 等待服务器启动
	time.Sleep(1 * time.Second)

	// 清理 Redis 队列中的残留任务
	Instance.ClearQueue("default")

	// 创建一个 5 秒后触发的 reminder 任务
	remindID := 12345
	payload, _ := json.Marshal(ReminderTaskPayload{RemindID: remindID})
	task := asynq.NewTask(TaskTypeReminder, payload)
	taskID := uniqueTaskID("reminder-test")

	// 入队延迟任务 (5秒后执行) - 使用 ProcessIn 选项
	info, err := Instance.Client().Enqueue(task, asynq.TaskID(taskID), asynq.ProcessIn(5*time.Second))
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, TaskTypeReminder, info.Type)
	t.Logf("Task enqueued: %s, will execute in 5 seconds", info.ID)

	// 检查任务状态
	taskInfo, err := Instance.GetTaskInfoByID("default", taskID)
	if err != nil {
		t.Logf("Failed to get task info: %v", err)
	} else {
		t.Logf("Task state: %s", taskInfo.State)
	}

	// 等待任务执行 (最多等待 15 秒)
	assert.Eventually(t, func() bool {
		return reminderExecuted.Load()
	}, 15*time.Second, 200*time.Millisecond, "Reminder 任务应该在 5 秒后执行")

	t.Logf("Reminder task executed successfully!")
}
