package services

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// 扇出队列与死信队列。写扩散是唯一真正需要异步化的链路：
// 一次发布要写到 N 个粉丝的时间线，N 随粉丝数增长，必须移出主请求。
const (
	fanoutQueue = "feed.fanout"
	fanoutDLQ   = "feed.fanout.dlq"
)

// fanoutTask 一条扇出任务：把某个视频写进作者的粉丝时间线
type fanoutTask struct {
	VideoID  uint `json:"video_id"`
	AuthorID uint `json:"author_id"`
}

// mqAddr 读取 MQ 地址，未设置时用本地默认值
func mqAddr() string {
	if v := os.Getenv("FEED_MQ_ADDR"); v != "" {
		return v
	}
	return "amqp://guest:guest@127.0.0.1:5672/"
}

var (
	mqConn  *amqp.Connection
	mqChan  *amqp.Channel
	mqReady bool
)

// SetupFanoutMQ 建立连接、声明队列与死信队列，并启动消费者。
// 任一步失败都返回错误，调用方据此把写扩散退回同步执行。
func SetupFanoutMQ() error {
	if os.Getenv("FEED_MQ_DISABLED") == "1" {
		return errors.New("消息队列已关闭")
	}

	conn, err := amqp.Dial(mqAddr())
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	// 失败的消息转入死信队列，避免在业务队列里反复重投
	if _, err := ch.QueueDeclare(fanoutQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": fanoutDLQ,
	}); err != nil {
		conn.Close()
		return err
	}
	if _, err := ch.QueueDeclare(fanoutDLQ, true, false, false, false, nil); err != nil {
		conn.Close()
		return err
	}

	deliveries, err := ch.Consume(fanoutQueue, "feed-fanout-worker", false, false, false, false, nil)
	if err != nil {
		conn.Close()
		return err
	}

	mqConn, mqChan, mqReady = conn, ch, true
	go consumeFanout(deliveries)
	return nil
}

// publishFanout 投递一条扇出任务；未就绪或投递失败时返回错误，调用方退回同步扇出
func publishFanout(authorID, videoID uint) error {
	if !mqReady {
		return errors.New("消息队列未就绪")
	}
	body, err := json.Marshal(fanoutTask{VideoID: videoID, AuthorID: authorID})
	if err != nil {
		return err
	}
	return mqChan.Publish("", fanoutQueue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         body,
	})
}

// consumeFanout 消费扇出任务。消息格式错误直接进死信；业务处理本身可重入，
// 因为写时间线用的是 ZADD，同一个视频重复写只是覆盖同一个分值。
func consumeFanout(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		var task fanoutTask
		if err := json.Unmarshal(d.Body, &task); err != nil {
			_ = d.Nack(false, false)
			continue
		}
		fanoutVideo(task.AuthorID, task.VideoID)
		_ = d.Ack(false)
	}
}
