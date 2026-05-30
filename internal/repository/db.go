package repository

import (
	"log"

	"github.com/zhanghanzhao/trade-hub/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.DexSwap{},
		&model.DexPool{},
		&model.Asset{},
		&model.CexOrder{},
	); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	return db
}
