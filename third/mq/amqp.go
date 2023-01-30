package mq

import (
	"strconv"
	"wopi-server/config"
	"wopi-server/g"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var (
	Conn      *amqp.Connection
	producers = map[exchangeName]producer{}
)

type exchangeType string

// exchange type enum
const (
	Topic  exchangeType = "topic"
	Direct exchangeType = "direct"
	Fanout exchangeType = "fanout"
)

type exchangeName string

// exchange name
const (
	YjGps exchangeName = "yj-gps"
)

type producer struct {
	exchangeName exchangeName
	channel      *amqp.Channel
}

func (p *producer) publishMsg(key string, mandatory, immediate bool, msg amqp.Publishing) {
	err := p.channel.Publish(string(p.exchangeName), key, mandatory, immediate, msg)
	failOnError(err, "publish msg fail, exchange = "+string(p.exchangeName))
}

//Connect 1. 尝试连接RabbitMQ，建立连接
// 该连接抽象了套接字连接，并为我们处理协议版本协商和认证等。
func Connect(config *config.AmpqConfig) {
	if !config.Enable {
		g.Log.Warn("ampq消息队列未启用，跳过连接")
		return
	}
	url := "amqp://" + config.Username + ":" + config.Password + "@" + config.Host + ":" + strconv.Itoa(config.Port) + "/"
	con, err := amqp.Dial(url)
	failOnError(err, "Failed to connect to RabbitMQ")
	Conn = con
}

//initChannel 2.创建一个通道
func initChannel() *amqp.Channel {
	ch, err := Conn.Channel()
	failOnError(err, "Failed to open a channel")
	return ch
}

//InitProducer 3.初始化生产者
func InitProducer() {
	producers = map[exchangeName]producer{}
	//云镜GPS生产者
	producers[YjGps] = producer{YjGps, initChannel()}
	err := producers[YjGps].channel.ExchangeDeclare(string(YjGps), string(Topic), true, false, false, false, nil)
	failOnError(err, "init yj gps producer fail")
	//
}

//Publish 生产消息
func Publish(exchange exchangeName, key string, mandatory, immediate bool, msg amqp.Publishing) {
	p := producers[exchange]
	go p.publishMsg(key, mandatory, immediate, msg)
}

//PublishJson 生产json消息
func PublishJson(exchange exchangeName, key string, payload []byte) {
	Publish(exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	})
}

//打印错误日志
func failOnError(err error, msg string) {
	if err != nil {
		zap.L().Error(msg, zap.Any("error", err))
		//重新连接mq
		Connect(config.Bean.Ampq)
		InitProducer()
	}
}
