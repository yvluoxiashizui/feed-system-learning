package services

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// 扇出队列与死信队列。写扩散是唯一需要异步化的链路：
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

// declareFanoutQueues 声明业务队列与死信队列，失败的消息转入死信，不在业务队列里反复重投
func declareFanoutQueues(ch *amqp.Channel) error {
	if _, err := ch.QueueDeclare(fanoutQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": fanoutDLQ,
	}); err != nil {
		return err
	}
	_, err := ch.QueueDeclare(fanoutDLQ, true, false, false, false, nil)
	return err
}

// SetupFanoutPublisher 建立投递侧连接，供 API 进程使用；失败时调用方退回同步扇出
func SetupFanoutPublisher() error {
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
	if err := declareFanoutQueues(ch); err != nil {
		conn.Close()
		return err
	}

	mqConn, mqChan, mqReady = conn, ch, true
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

// ConsumeFanout 以阻塞方式消费扇出任务，供 Worker 进程使用。
// 连接断开时返回，由调用方重连；消息格式错误直接进死信，业务处理本身可重入。
func ConsumeFanout() error {
	conn, err := amqp.Dial(mqAddr())
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	if err := declareFanoutQueues(ch); err != nil {
		return err
	}

	deliveries, err := ch.Consume(fanoutQueue, "feed-fanout-worker", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for d := range deliveries {
		var task fanoutTask
		if err := json.Unmarshal(d.Body, &task); err != nil {
			_ = d.Nack(false, false)
			continue
		}
		fanoutVideo(task.AuthorID, task.VideoID)
		_ = d.Ack(false)
	}
	return nil
}
