package model

import (
	"aibuddy/pkg/config"
	"database/sql/driver"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

const defaultTableNamePrefix = "buddy_"

// DefaultDomainName 默认域名
const DefaultDomainName = "https://ai.ipai.fans"

// LocalTime 自定义时间类型，JSON序列化为 "2006-01-02 15:04:05" 格式
type LocalTime time.Time

// MarshalJSON 实现 json.Marshaler 接口
func (t LocalTime) MarshalJSON() ([]byte, error) {
	formatted := `"` + time.Time(t).Format("2006-01-02 15:04:05") + `"`
	return []byte(formatted), nil
}

// Value 实现 driver.Valuer 接口，用于数据库写入
func (t LocalTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (t *LocalTime) Scan(value any) error {
	if value == nil {
		*t = LocalTime(time.Time{})
		return nil
	}
	if v, ok := value.(time.Time); ok {
		*t = LocalTime(v)
	}
	return nil
}

// Unix 返回 Unix 时间戳
func (t LocalTime) Unix() int64 {
	return time.Time(t).Unix()
}

// Time 返回标准 time.Time
func (t LocalTime) Time() time.Time {
	return time.Time(t)
}

// TableName returns the table name with the default prefix.
func TableName(name string) string {
	return defaultTableNamePrefix + name
}

// DB represents a database connection.
type DB struct {
	db *gorm.DB
}

// Conn returns a new database connection.
func Conn() *DB {
	cfg := config.Instance.Storage.Database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().In(time.Local)
		},
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束
	})
	if err != nil {
		panic(err)
	}
	return &DB{db: db}
}

// GetDB returns the underlying GORM database instance.
func (d *DB) GetDB() *gorm.DB {
	return d.db
}

// GenerateQuery generates query code for GORM models.
func (d *DB) GenerateQuery() {
	configPath := config.FoundConfigPath()

	g := gen.NewGenerator(gen.Config{
		OutPath: filepath.Join(configPath, "../internal/query"),
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	models := []any{
		User{},
		Agent{},
		Device{},
		Reminder{},
		AnniversaryReminder{},

		DeviceSN{},
		DeviceInfo{},

		// OTA
		DeviceOta{},
		OtaResource{},

		ChatDialogue{},

		DeviceRelationship{},

		DeviceMessage{},

		NFC{},

		UserAgent{},

		Emotion{},

		PomodoroClock{},

		GrowthReport{},

		Feedback{},
	}

	_ = d.db.AutoMigrate(models...)

	g.UseDB(d.db)

	g.ApplyBasic(models...)

	g.Execute()
}

// ExtractFilename 从嵌套的URL中提取最终的filename值
// 例如: https://ai.ipai.fans/api/v1/file/...?filename=30:ED:A0:E9:F3:22/9687183842.mp3
// 返回: 30:ED:A0:E9:F3:22/9687183842.mp3
func ExtractFilename(url string) string {
	if url == "" {
		return url
	}
	for strings.Contains(url, "filename=") {
		idx := strings.LastIndex(url, "filename=")
		if idx == -1 {
			break
		}
		value := url[idx+9:] // len("filename=") = 9
		if strings.Contains(value, "filename=") {
			url = value
			continue
		}
		return value
	}
	return url
}
