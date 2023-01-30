package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Bean *Config

// 默认配置文件信息
var tomlConf = []byte(`
[ampq]
enable = false
host = "10.90.11.108"
password = "test"
port = 7256
username = "test"

[logger]
jsonformat = false
loginconsole = true
loglevel = "debug"
logpath = "./demo.log"
maxage = 5  #最大保存天数
maxbackups = 30  #最大备份数量
maxsize = 10  #日志文件最大大小
showline = true
compress = false #是否启用压缩

[postgres]
enable = false                      #是否启动postgresql数据库连接
dbname = "cloud_command2"
host = "10.90.11.108"
maxidleconn = 5
maxopenconn = 10
password = "f420e79b152d4498affe3f50c065e0c3"
port = 5432
user = "postgres"

[mysql]
enable = false                      #是否启动mysql数据库连接
host = "10.90.10.69"
port = 3306
dbname = "vdoor"
user = "gosun"
password = "video"
maxidleconn = 5
maxopenconn = 10
properties = "parseTime=true&loc=Local"

[server]
port = 33597                               #访问端口
localFileRootdir = "D:/download/caddy"     #本地文件目录
wopiServer = "http://10.90.11.233:9980"    #wopi server
context = "/go-iris"                       #应用的上下文,已 '/'开始
token = "123456"                           #密码（用于编辑）
`)

// Config config settings
type Config struct {
	Server   *ServerConfig
	Logger   *LogConfig
	Ampq     *AmpqConfig
	Postgres *PostgresConfig
	Mysql    *MysqlConfig
}

type ServerConfig struct {
	Port             int    //Rest端口
	LocalFileRootdir string //本地文件目录
	WopiServer       string //wopi server
	Context          string //应用的上下文,已 '/'开始
	Token            string //编辑模式token
}

// LogConfig logger config entity
type LogConfig struct {
	LogPath      string `json:"logPath"`
	LogLevel     string `json:"logLevel"`
	MaxSize      int    `json:"maxSize"`
	MaxBackups   int    `json:"maxBackups"`
	MaxAge       int    `json:"maxAge"`
	Compress     bool   `json:"compress"`
	JsonFormat   bool   `json:"jsonFormat"`
	ShowLine     bool   `json:"showLine"`
	LogInConsole bool   `json:"logInConsole"`
}

// AmpqConfig mq config entity
type AmpqConfig struct {
	Enable   bool
	Host     string
	Port     int
	Username string
	Password string
}

// PostgresConfig postgresql  config entity
type PostgresConfig struct {
	Enable      bool
	Host        string
	Port        int32
	Dbname      string
	User        string
	Password    string
	MaxOpenConn int
	MaxIdleConn int
}

// MysqlConfig mysql config
type MysqlConfig struct {
	Enable      bool
	Host        string
	Port        int32
	Dbname      string
	User        string
	Password    string
	MaxOpenConn int
	MaxIdleConn int
	Properties  string
}

func LoadConfig() {
	confDir := "./conf"
	confName := "conf"
	confType := "toml"
	viper.SetConfigName(confName)
	viper.SetConfigType(confType)
	viper.AddConfigPath(confDir)
	if err := viper.ReadConfig(bytes.NewBuffer(tomlConf)); err != nil {
		log.Fatal(fmt.Sprintf("init config failed : %v", err))
	}
	//判断配置目录是否存在
	exists, _ := PathExists(confDir)
	if !exists {
		if err := os.Mkdir(confDir, os.ModePerm); err != nil {
			log.Fatalf("mk config dir failed  %v ", err)
		}
	}
	confExists, _ := PathExists(confDir + "/" + confName + "." + confType)
	if !confExists {
		if err := viper.SafeWriteConfig(); err != nil {
			log.Printf("write config failed: %v", err)
		}
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("read config failed : ", zap.Any(" ", err))
	}
	if err := viper.Unmarshal(&Bean); err != nil {
		log.Fatal("unmarshal config file failed : ", zap.Any(" ", err))
	}
	jsonStr, err := json.MarshalIndent(Bean, "", "\t")
	if err != nil {
		log.Fatalf("config json conv error %v", err)
	}
	log.Printf("config info :%v", string(jsonStr))
}

// PathExists : dir or file exist
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { // 文件或者目录存在
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
