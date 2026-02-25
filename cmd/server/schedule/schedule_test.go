package schedule

import (
	"aibuddy/pkg/config"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	config.Instance = &config.Config{
		Storage: &config.StorageConfig{
			Redis: &config.RedisConfig{
				Host:     "localhost",
				Port:     6379,
				DB:       0,
				Password: "123456",
			},
		},
	}
	m.Run()
}

func Test_StartSchedule(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Second)
	defer cancel()

	err := StartSchedule(ctx)

	t.Logf("StartSchedule result: %v", err)
}

func Test_DemoDelay(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Second)
	defer cancel()
	go func() {
		err := StartSchedule(ctx)
		assert.NoError(t, err)
	}()

	time.Sleep(time.Second * 1)
	// 发送延迟任务
	demoDelay := &DemoDelay{}
	err := WrapDelayedTask(demoDelay, []TaskArgs{
		{
			Name:  "name",
			Type:  "string",
			Value: "apple",
		},
		{
			Name:  "age",
			Type:  "int",
			Value: 18,
		},
	}...)
	assert.NoError(t, err)

	time.Sleep(time.Until(demoDelay.Eta()) + time.Second*3)
}

func Test_TimeParse(_ *testing.T) {
	gen := NewCronGenerator()

	// 示例 1: 每天下午 2 点
	t1, _ := time.Parse("15:04", "14:00")
	cron1, _ := gen.GenerateCron(Daily, t1, 0, 0)
	fmt.Printf("每天下午2点: %s\n", cron1) // 0 14 * * *

	// 示例 2: 每周一上午 9 点
	t2, _ := time.Parse("15:04", "09:00")
	cron2, _ := gen.GenerateCron(Weekly, t2, 1, 0) // 1 = 周一
	fmt.Printf("每周一上午9点: %s\n", cron2)             // 0 9 * * 1

	// 示例 3: 每月 15 号下午 3 点
	t3, _ := time.Parse("15:04", "15:00")
	cron3, _ := gen.GenerateCron(Monthly, t3, 0, 15)
	fmt.Printf("每月15号下午3点: %s\n", cron3) // 0 15 15 * *

	// 示例 4: 一次性提醒 - 2024年12月25日上午10点
	t4, _ := time.Parse("2006-01-02 15:04", "2024-12-25 10:00")
	cron4, _ := gen.GenerateCron(Once, t4, 0, 0)
	fmt.Printf("2024-12-25 上午10点: %s\n", cron4) // 0 10 25 12 *

	// 示例 5: 使用字符串解析
	cron5, _ := gen.GenerateCronFromString(Daily, "14:30", 0, 0)
	fmt.Printf("每天14:30: %s\n", cron5) // 30 14 * * *
}
