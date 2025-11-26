package cachedb

import (
	"fmt"
	"log"
	"os"

	"github.com/fireinrain/javbus-api/config"
	"github.com/fireinrain/javbus-api/utils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var CacheDb *gorm.DB

func InitDataBase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.DBType {
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.DBServerPath), &gorm.Config{})
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.DBServerPath), &gorm.Config{})
	case "sqlite":
		// SQLite 文件判断
		if _, statErr := os.Stat(cfg.DBServerPath); os.IsNotExist(statErr) {
			// 文件不存在，创建目录
			if dir := utils.GetFileDir(cfg.DBServerPath); dir != "" {
				if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
					return nil, fmt.Errorf("failed to create sqlite directory: %v", mkErr)
				}
			}
			log.Printf("SQLite 文件不存在，已创建新文件：%s\n", cfg.DBServerPath)
		} else {
			log.Printf("SQLite 文件已存在：%s\n", cfg.DBServerPath)
		}
		db, err = gorm.Open(sqlite.Open(cfg.DBServerPath), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported db type: %s", cfg.DBType)
	}

	if err != nil {
		return nil, err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// 自动迁移 TODO
	//if err := db.AutoMigrate(&models.User{}); err != nil {
	//	return nil, err
	//}

	CacheDb = db
	return db, nil
}
