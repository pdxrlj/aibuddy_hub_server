package model

import (
	"aibuddy/pkg/config"
	"fmt"
	"path/filepath"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

const defaultTableNamePrefix = "buddy_"

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
	}

	_ = d.db.AutoMigrate(models...)

	g.UseDB(d.db)

	g.ApplyBasic(models...)

	g.Execute()
}
