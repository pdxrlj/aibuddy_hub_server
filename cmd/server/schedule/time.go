package schedule

import (
	"fmt"
	"time"
)

// ReminderType represents the type of reminder.
type ReminderType string

const (
	// Daily represents a daily reminder.
	Daily ReminderType = "daily"
	// Weekly represents a weekly reminder.
	Weekly ReminderType = "weekly"
	// Monthly represents a monthly reminder.
	Monthly ReminderType = "monthly"
	// Once represents a one-time reminder.
	Once ReminderType = "once"
)

// CronGenerator generates cron expressions for reminders.
type CronGenerator struct{}

// NewCronGenerator creates a new CronGenerator.
func NewCronGenerator() *CronGenerator {
	return &CronGenerator{}
}

// GenerateCron generates a cron expression.
// reminderType: 提醒类型（daily/weekly/monthly/once）
// t: 时间
// weekday: 星期几（0-6，0=周日），仅 weekly 时需要
// dayOfMonth: 每月第几天（1-31），仅 monthly 时需要
func (cg *CronGenerator) GenerateCron(reminderType ReminderType, t time.Time, weekday int, dayOfMonth int) (string, error) {
	hour := t.Hour()
	minute := t.Minute()

	switch reminderType {
	case Daily:
		// 每天指定时间
		return fmt.Sprintf("%d %d * * *", minute, hour), nil

	case Weekly:
		// 每周指定星期几的指定时间
		if weekday < 0 || weekday > 6 {
			return "", fmt.Errorf("weekday must be between 0-6")
		}
		return fmt.Sprintf("%d %d * * %d", minute, hour, weekday), nil

	case Monthly:
		// 每月指定日期的指定时间
		if dayOfMonth < 1 || dayOfMonth > 31 {
			return "", fmt.Errorf("dayOfMonth must be between 1-31")
		}
		return fmt.Sprintf("%d %d %d * *", minute, hour, dayOfMonth), nil

	case Once:
		// 一次性提醒（指定具体日期时间）
		return fmt.Sprintf("%d %d %d %d *", minute, hour, t.Day(), int(t.Month())), nil

	default:
		return "", fmt.Errorf("unknown reminder type: %s", reminderType)
	}
}

// GenerateCronFromString generates a cron expression from a time string.
func (cg *CronGenerator) GenerateCronFromString(reminderType ReminderType, timeStr string, weekday int, dayOfMonth int) (string, error) {
	// 支持多种时间格式
	formats := []string{
		"15:04",            // 24小时制：14:30
		"3:04 PM",          // 12小时制：2:30 PM
		"下午3:04",           // 中文：下午3:04
		"2006-01-02 15:04", // 完整日期时间
	}

	var t time.Time
	var err error

	for _, format := range formats {
		t, err = time.Parse(format, timeStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return "", fmt.Errorf("无法解析时间格式: %s", timeStr)
	}

	return cg.GenerateCron(reminderType, t, weekday, dayOfMonth)
}
