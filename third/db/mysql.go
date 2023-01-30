package db

import (
	"fmt"
	"time"
	"wopi-server/g"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"wopi-server/config"
)

var Mysqldb *gorm.DB

func OpenMysql(conf *config.MysqlConfig) *gorm.DB {
	if !conf.Enable {
		g.Log.Warn("MySQL数据库未启用，跳过连接")
		return nil
	}
	dbUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", conf.User, conf.Password, conf.Host, conf.Port, conf.Dbname, conf.Properties)
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	// dbUrl := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err1 := gorm.Open(mysql.Open(dbUrl), &gorm.Config{})
	if err1 != nil {
		zap.L().Panic("mysql connect error", zap.Any("", err1))
	}
	dbs, err2 := db.DB()
	if err2 != nil {
		zap.L().Error("db ping error ", zap.Any("postgresql : ", err2))
	}
	if err3 := dbs.Ping(); err3 != nil {
		zap.L().Error("db ping error ", zap.Any("postgresql : ", err3))
	}
	if conf.MaxOpenConn < 10 {
		conf.MaxOpenConn = 10
	}
	if conf.MaxIdleConn < 5 {
		conf.MaxIdleConn = 5
	}
	dbs.SetMaxOpenConns(conf.MaxOpenConn)
	dbs.SetMaxIdleConns(conf.MaxIdleConn)
	dbs.SetConnMaxLifetime(time.Hour)
	return db
}
