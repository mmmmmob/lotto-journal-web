package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// PoolConfig holds the connection pool tuning configurations.
type PoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// ConnectDatabase establishes the connection to Postgres via GORM, applies connection pool optimizations, and returns the DB instance.
func ConnectDatabase(dsn string, poolCfg PoolConfig) (*gorm.DB, error) {
	var err error

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	maxIdle := poolCfg.MaxIdleConns
	maxOpen := poolCfg.MaxOpenConns
	maxLifetime := poolCfg.ConnMaxLifetime
	maxIdleTime := poolCfg.ConnMaxIdleTime

	// Validate and clamp parameters defensively inside ConnectDatabase to ensure safe limits
	if maxIdle < 0 {
		maxIdle = 0
	}
	if maxOpen < 1 {
		maxOpen = 5 // Safe default fallback (matching config.go) to prevent unlimited connection leaks
	}
	if maxLifetime < 30*time.Second {
		maxLifetime = 3 * time.Minute // Safe default fallback (matching config.go) to prevent infinite lifetime connection leaks
	}
	if maxIdleTime < 10*time.Second {
		maxIdleTime = 1 * time.Minute // Safe default fallback (matching config.go) to prevent infinite idle connection leaks
	}

	// Optimize connection settings for serverless database (Neon) and scale-to-zero autoscaling (Fly.io)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetConnMaxLifetime(maxLifetime)
	sqlDB.SetConnMaxIdleTime(maxIdleTime)

	log.Printf("Database connection successful (pool config: maxOpen=%d, maxIdle=%d, maxLifetime=%v, maxIdleTime=%v)",
		maxOpen, maxIdle, maxLifetime, maxIdleTime)

	DB = db // Populate global for backward compatibility
	return db, nil
}
