package schedule

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/RichardKnop/logging"
	mlog "github.com/RichardKnop/machinery/v2/log"
)

// slogAdapter 将 slog 适配到 machinery 的 LoggerInterface
type slogAdapter struct {
	logger *slog.Logger
}

func (s *slogAdapter) Print(v ...interface{}) {
	s.logger.Info(fmt.Sprint(v...))
}

func (s *slogAdapter) Printf(f string, v ...interface{}) {
	s.logger.Info(fmt.Sprintf(f, v...))
}

func (s *slogAdapter) Println(v ...interface{}) {
	s.logger.Info(fmt.Sprint(v...))
}

func (s *slogAdapter) Fatal(v ...interface{}) {
	log.Fatal(v...)
}

func (s *slogAdapter) Fatalf(f string, v ...interface{}) {
	log.Fatalf(f, v...)
}

func (s *slogAdapter) Fatalln(v ...interface{}) {
	log.Fatalln(v...)
}

func (s *slogAdapter) Panic(v ...interface{}) {
	log.Panic(v...)
}

func (s *slogAdapter) Panicf(f string, v ...interface{}) {
	log.Panicf(f, v...)
}

func (s *slogAdapter) Panicln(v ...interface{}) {
	log.Panicln(v...)
}

var _ logging.LoggerInterface = (*slogAdapter)(nil)

func init() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // 设置日志级别
	}))

	adapter := &slogAdapter{logger: logger}
	mlog.Set(adapter)
}
