package service

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQBuildQueue 使用 RabbitMQ 实现构建队列
type RabbitMQBuildQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// NewRabbitMQBuildQueue 使用连接串与队列名初始化队列
func NewRabbitMQBuildQueue(url, queue string) (*RabbitMQBuildQueue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}
	return &RabbitMQBuildQueue{conn: conn, channel: ch, queue: queue}, nil
}

// Enqueue 向 RabbitMQ 发布构建任务
func (q *RabbitMQBuildQueue) Enqueue(ctx context.Context, job BuildJob) error {
	body := fmt.Sprintf("%d|%d|%d|%s|%s", job.BuildID, job.PipelineID, job.ProjectID, job.CommitSHA, job.Branch)
	// 简单字符串编码；生产环境建议使用 JSON
	return q.channel.PublishWithContext(ctx, "", q.queue, false, false, amqp.Publishing{
		ContentType:  "text/plain",
		Body:         []byte(body),
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
	})
}

// Close 关闭通道与连接
func (q *RabbitMQBuildQueue) Close() {
	if q.channel != nil {
		_ = q.channel.Close()
	}
	if q.conn != nil {
		_ = q.conn.Close()
	}
}

// SimpleLocalExecutor 从队列消费任务并记录执行占位日志
type SimpleLocalExecutor struct{}

func NewSimpleLocalExecutor() *SimpleLocalExecutor { return &SimpleLocalExecutor{} }

// Execute 执行占位实现，未来用于运行流水线编排
func (e *SimpleLocalExecutor) Execute(ctx context.Context, job BuildJob) error {
	log.Printf("Executing build job: build_id=%d pipeline_id=%d project_id=%d commit=%s branch=%s", job.BuildID, job.PipelineID, job.ProjectID, job.CommitSHA, job.Branch)
	return nil
}
