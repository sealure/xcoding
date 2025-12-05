package db

import (
    "context"
    "fmt"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// GormDB 是对 gorm.DB 的轻量封装，统一连接与关闭语义
type GormDB struct {
    DB *gorm.DB
}

// NewGorm 使用 Postgres 驱动创建新的 GORM 数据库连接
func NewGorm(databaseURL string) (*GormDB, error) {
    db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, fmt.Errorf("unable to connect to database with GORM: %w", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("unable to get underlying sql.DB: %w", err)
    }
    sqlDB.SetMaxIdleConns(2)
    sqlDB.SetMaxOpenConns(10)
    sqlDB.SetConnMaxLifetime(time.Hour)

    return &GormDB{DB: db}, nil
}

// Close 关闭底层连接池
func (database *GormDB) Close() error {
    sqlDB, err := database.DB.DB()
    if err != nil {
        return fmt.Errorf("unable to get underlying sql.DB: %w", err)
    }
    return sqlDB.Close()
}

// BeginTx 开始事务（携带上下文）
func (database *GormDB) BeginTx(ctx context.Context) *gorm.DB { return database.DB.WithContext(ctx).Begin() }

// GetDB 暴露原生 *gorm.DB
func (database *GormDB) GetDB() *gorm.DB { return database.DB }

// AutoMigrate 自动迁移模型
func (database *GormDB) AutoMigrate(models ...interface{}) error { return database.DB.AutoMigrate(models...) }