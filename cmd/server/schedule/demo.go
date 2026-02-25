package schedule

import (
	"fmt"
	"log/slog"
	"time"
)

var _ Schedule = &Demo{}

// Demo is a demo schedule for periodic tasks.
type Demo struct{}

var _ = Register(&Demo{})

// Spec returns the cron specification for the demo task.
func (d *Demo) Spec() string {
	// 设置14点47分执行
	// return "@every 3s"
	return "58 14 * * *"
}

// Eta returns the estimated time of arrival for the task.
func (d *Demo) Eta() time.Time {
	return time.Time{}
}

// Name returns the name of the demo task.
func (d *Demo) Name() string {
	return "test_periodic"
}

// Task returns the task function for the demo.
func (d *Demo) Task() func(args ...interface{}) error {
	return func(_ ...interface{}) error {
		slog.Info("demo task running")

		return nil
	}
}

var _ Schedule = &DemoDelay{}

// DemoDelay is a demo schedule for delayed tasks.
type DemoDelay struct{}

var _ = Register(&DemoDelay{})

// Spec returns the cron specification for the delayed demo task.
func (d *DemoDelay) Spec() string {
	return ""
}

// Eta returns the estimated time of arrival for the delayed task.
func (d *DemoDelay) Eta() time.Time {
	return time.Now().Add(time.Second * 5)
}

// Name returns the name of the delayed demo task.
func (d *DemoDelay) Name() string {
	return "test_delay"
}

// Task returns the task function for the delayed demo.
func (d *DemoDelay) Task() func(args ...any) error {
	slog.Info("start time", "time", time.Now().Format(time.DateTime))
	return func(_ ...any) error {
		slog.Info("demo delay task running", "args", fmt.Sprintf("%v", ""), "current time", time.Now().Format(time.DateTime))
		return nil
	}
}
