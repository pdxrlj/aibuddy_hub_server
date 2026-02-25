package schedule

import (
	"aibuddy/pkg/schedule"
	"errors"
	"reflect"
	"time"

	"github.com/RichardKnop/machinery/v2/tasks"
)

// Schedule defines the interface for scheduled tasks.
type Schedule interface {
	Spec() string // @every 1m
	Eta() time.Time
	Name() string                          // test_periodic
	Task() func(args ...interface{}) error // 任务函数
}

var _Schedules = []Schedule{}

// Register registers a new schedule.
func Register(scheduleI Schedule) error {
	_Schedules = append(_Schedules, scheduleI)
	return nil
}

// getDefaultArgs 使用反射获取函数参数默认值
func getDefaultArgs(fn interface{}) []tasks.Arg {
	fnType := reflect.TypeOf(fn)

	if fnType.Kind() == reflect.Ptr {
		fnType = fnType.Elem()
	}

	if fnType.IsVariadic() {
		return []tasks.Arg{}
	}

	args := make([]tasks.Arg, fnType.NumIn())
	for i := 0; i < fnType.NumIn(); i++ {
		argType := fnType.In(i)
		args[i] = tasks.Arg{
			Type:  argType.String(),
			Value: getDefaultValue(argType),
		}
	}
	return args
}

// getDefaultValue 根据类型返回默认值
func getDefaultValue(t reflect.Type) interface{} {
	switch t.Kind() {
	case reflect.String:
		return ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return 0
	case reflect.Float32, reflect.Float64:
		return 0.0
	case reflect.Bool:
		return false
	case reflect.Slice:
		return []interface{}{}
	case reflect.Map:
		return map[string]interface{}{}
	default:
		return nil
	}
}

// RegisterTasks registers periodic tasks.
func RegisterTasks() error {
	for _, scheduleI := range _Schedules {
		spec := scheduleI.Spec()
		name := scheduleI.Name()
		taskFn := scheduleI.Task()

		_ = schedule.Instance.Server.RegisterTask(name, taskFn)

		args := getDefaultArgs(taskFn)
		if spec != "" {
			_ = schedule.Instance.Server.RegisterPeriodicTask(spec, name, &tasks.Signature{
				Name: name,
				Args: args,
			})
		}
	}

	return nil
}

// TaskArgs defines arguments for a delayed task.
type TaskArgs struct {
	Name  string
	Type  string
	Value interface{}
}

// WrapDelayedTask wraps a delayed task with its arguments.
func WrapDelayedTask(taskFn Schedule, taskArgs ...TaskArgs) error {
	name := taskFn.Name()
	eta := taskFn.Eta()
	if eta.IsZero() {
		return errors.New("eta is zero")
	}

	var args []tasks.Arg
	if len(taskArgs) > 0 {
		args = make([]tasks.Arg, len(taskArgs))
		for i, arg := range taskArgs {
			args[i] = tasks.Arg{
				Name:  arg.Name,
				Type:  arg.Type,
				Value: arg.Value,
			}
		}
	} else {
		args = getDefaultArgs(taskFn.Task())
	}

	signature := &tasks.Signature{
		Name: name,
		Args: args,
		ETA:  &eta,
	}

	_, err := schedule.Instance.Server.SendTask(signature)

	return err
}
