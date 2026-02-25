// Package schedule provides task scheduling functionality using machinery.
package schedule

import (
	"errors"
	"fmt"
	"sync"

	"github.com/RichardKnop/machinery/v2"
	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	"github.com/RichardKnop/machinery/v2/config"
	eagerlock "github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/RichardKnop/machinery/v2/tasks"
)

// Instance is the global schedule instance.
var Instance *Schedule

// Schedule manages task scheduling using machinery.
type Schedule struct {
	Server  *machinery.Server
	worker  *machinery.Worker
	closing chan struct{}
	mu      sync.Mutex

	ErrorHandler func(err error)

	PreTaskHandler  func(signature *tasks.Signature)
	PostTaskHandler func(signature *tasks.Signature)
}

// Config holds the configuration for the schedule.
type Config struct {
	RedisDB       int
	RedisHost     string
	RedisPort     int
	RedisPassword string
}

// New creates a new Schedule instance.
func New(scheduleConfig *Config) (*Schedule, error) {
	if scheduleConfig == nil {
		return nil, errors.New("schedule config is nil")
	}
	cnf := &config.Config{
		DefaultQueue:    "machinery_tasks",
		ResultsExpireIn: 3600,
		Redis: &config.RedisConfig{
			MaxIdle:                3,
			IdleTimeout:            240,
			ReadTimeout:            15,
			WriteTimeout:           15,
			ConnectTimeout:         15,
			NormalTasksPollPeriod:  1000,
			DelayedTasksPollPeriod: 500,
			DelayedTasksKey:        "aibuddy_task:",
		},
	}

	redisHost := scheduleConfig.RedisHost
	redisPort := scheduleConfig.RedisPort
	redisPassword := scheduleConfig.RedisPassword
	redisDB := scheduleConfig.RedisDB
	redisAddr := fmt.Sprintf("%s:%d", redisHost, redisPort)

	broker := redisbroker.New(cnf, redisAddr, "", redisPassword, "", redisDB)
	backend := redisbackend.New(cnf, redisAddr, "", redisPassword, "", redisDB)
	lock := eagerlock.New()
	server := machinery.NewServer(cnf, broker, backend, lock)

	Instance = &Schedule{Server: server}
	return Instance, nil
}

// Consumer 消费任务
func (s *Schedule) Consumer() error {
	s.closing = make(chan struct{})
	worker := s.Server.NewWorker("aibuddy", 0)
	s.worker = worker

	if s.ErrorHandler != nil {
		worker.SetErrorHandler(s.ErrorHandler)
	}
	if s.PreTaskHandler != nil {
		worker.SetPreTaskHandler(s.PreTaskHandler)
	}
	if s.PostTaskHandler != nil {
		worker.SetPostTaskHandler(s.PostTaskHandler)
	}

	if err := worker.Launch(); err != nil {
		return err
	}

	return nil
}

// Shutdown 关闭
func (s *Schedule) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.worker != nil {
		defer func() {
			s.worker = nil
		}()

		func() {
			defer func() {
				_ = recover()
			}()
			s.worker.Quit()
		}()
	}
	if s.closing != nil {
		close(s.closing)
		s.closing = nil
	}
	return nil
}
