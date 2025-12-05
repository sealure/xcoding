package consumer

import (
	"context"
	"fmt"
	"strings"
	"time"
	"xcoding/apps/ci/executor_service/internal/executor"
	"xcoding/apps/ci/executor_service/internal/parser"
	"xcoding/apps/ci/executor_service/models"
	civ1 "xcoding/gen/go/ci/v1"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type BuildMessage struct{ BuildID uint64 }

type ExecutorClient interface {
	CancelBuild(ctx context.Context, in *civ1.CancelExecutorBuildRequest, opts ...grpc.CallOption) (*civ1.CancelExecutorBuildResponse, error)
}

type QueueConsumer struct {
	url    string
	queue  string
	conn   *amqp.Connection
	ch     *amqp.Channel
	client ExecutorClient
	db     *gorm.DB
	k8s    *executor.K8sEnv
}

func NewQueueConsumer(url, queue string, client ExecutorClient, db *gorm.DB, namespace string) *QueueConsumer {
	return &QueueConsumer{url: url, queue: queue, client: client, db: db}
}

func (c *QueueConsumer) Start(ctx context.Context) error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open channel: %w", err)
	}
	if _, err := ch.QueueDeclare(c.queue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("declare queue: %w", err)
	}
	c.conn, c.ch = conn, ch

	kenv, err := executor.NewK8sEnv()
	if err != nil {
		return fmt.Errorf("k8s env: %w", err)
	}
	c.k8s = kenv
	msgs, err := ch.Consume(c.queue, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}
	go func() {
		for m := range msgs {
			parts := strings.Split(string(m.Body), "|")
			if len(parts) < 1 {
				continue
			}
			var buildID uint64
			fmt.Sscanf(parts[0], "%d", &buildID)
			_ = c.handleBuild(ctx, buildID)
		}
	}()
	return nil
}

func (c *QueueConsumer) handleBuild(ctx context.Context, buildID uint64) error {
	if c.db != nil {
		now := time.Now()
		_ = c.db.Model(&models.Build{}).Where("id = ?", buildID).Updates(map[string]any{"status": int32(civ1.BuildStatus_BUILD_STATUS_RUNNING), "started_at": &now}).Error
	}
	var snap models.BuildSnapshot
	if err := c.db.Where("build_id = ?", buildID).First(&snap).Error; err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}
	wf, err := parser.ParseWorkflowYAML(snap.WorkflowYAML)
	if err != nil {
		return fmt.Errorf("parse workflow: %w", err)
	}

	// 延迟初始化：检查是否已有 BuildJob，若无则创建
	var count int64
	if err := c.db.Model(&models.BuildJob{}).Where("build_id = ?", buildID).Count(&count).Error; err == nil && count == 0 {
		idx := int32(0)
		for name, j := range wf.Jobs {
			idx++
			_ = c.db.Create(&models.BuildJob{BuildID: buildID, Name: name, Status: "pending", Index: idx}).Error
			for _, n := range j.Needs {
				_ = c.db.Create(&models.BuildJobEdge{BuildID: buildID, FromJob: n, ToJob: name}).Error
			}
			stepIdx := int32(0)
			for _, st := range j.Steps {
				stepIdx++
				_ = c.db.Create(&models.BuildStep{BuildID: buildID, JobName: name, Index: stepIdx, Name: st.Name, Status: "pending"}).Error
			}
		}
	}

	eng := executor.NewEngine(c.k8s, c.db, nil)
	err = eng.RunWorkflow(ctx, buildID, wf)
	return err
}

func (c *QueueConsumer) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}
