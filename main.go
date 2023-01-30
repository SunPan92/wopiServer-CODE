package main

import (
	"strconv"
	"wopi-server/config"
	"wopi-server/handler"
	"wopi-server/middleware"
	"wopi-server/third/cache"
	"wopi-server/third/db"

	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
	"gorm.io/gen"
)

func main() {
	//加载配置文件
	config.LoadConfig()
	//初始化日志
	middleware.InitLogger(config.Bean.Logger)
	//连接数据库
	db.Mysqldb = db.OpenMysql(config.Bean.Mysql)
	db.PgDb = db.OpenPg(config.Bean.Postgres)
	//初始化本地缓存
	cache.InitCache()
	//设置本地文件的根目录
	handler.Initial()
	//初始化router
	application := InitRouter()
	if err := application.Run(iris.Addr(":" + strconv.Itoa(config.Bean.Server.Port))); err != nil {
		zap.L().Fatal("initial router fail", zap.Any("error", err))
	}
}

// generate code
func generateGormStructCode() {
	// specify the output directory (default: "./query")
	// ### if you want to query without context constrain, set mode gen.WithoutContext ###
	g := gen.NewGenerator(gen.Config{
		OutPath: "../gen/query",
		/* Mode: gen.WithoutContext|gen.WithDefaultQuery*/
		//if you want the nullable field generation property to be pointer type, set FieldNullable true
		/* FieldNullable: true,*/
		//if you want to assign field which has default value in `Create` API, set FieldCoverable true, reference: https://gorm.io/docs/create.html#Default-Values
		FieldCoverable: true,
		//if you want to generate index tags from database, set FieldWithIndexTag true
		/* FieldWithIndexTag: true,*/
		//if you want to generate type tags from database, set FieldWithTypeTag true
		FieldWithTypeTag: false,
		//if you need unit tests for query code, set WithUnitTest true
		/* WithUnitTest: true, */
	})

	// reuse the database connection in Project or create a connection here
	// if you want to use GenerateModel/GenerateModelAs, UseDB is necessary or it will panic
	// db, _ := gorm.Open(mysql.Open("root:@(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local"))
	dbs := db.OpenPg(config.Bean.Postgres)

	g.UseDB(dbs)
	g.ApplyBasic(g.GenerateModelAs("file_info", "FileInfo"))
	// execute the action of code generation
	g.Execute()
	zap.L().Info("generate code success")
}
