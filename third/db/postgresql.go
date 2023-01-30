package db

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
	"wopi-server/g"

	"wopi-server/config"
)

var PgDb *gorm.DB

// OpenPg open pg connection
func OpenPg(pgConf *config.PostgresConfig) *gorm.DB {
	if !pgConf.Enable {
		g.Log.Warn("PostgreSQL数据库未启用，跳过连接")
		return nil
	}
	dsnUrl := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable", pgConf.Host, pgConf.Port, pgConf.User, pgConf.Dbname, pgConf.Password)
	db, err1 := gorm.Open(postgres.Open(dsnUrl), &gorm.Config{})
	if err1 != nil {
		zap.L().Error("db open error ", zap.Any("postgresql : ", err1))
	}
	dbs, err2 := db.DB()
	if err2 != nil {
		zap.L().Error("db ping error ", zap.Any("postgresql : ", err2))
	}
	if err3 := dbs.Ping(); err3 != nil {
		zap.L().Error("db ping error ", zap.Any("postgresql : ", err3))
	}
	if pgConf.MaxOpenConn < 10 {
		pgConf.MaxOpenConn = 10
	}
	if pgConf.MaxIdleConn < 5 {
		pgConf.MaxIdleConn = 5
	}
	dbs.SetMaxOpenConns(pgConf.MaxOpenConn)
	dbs.SetMaxIdleConns(pgConf.MaxIdleConn)
	dbs.SetConnMaxLifetime(time.Hour)
	return db
}
