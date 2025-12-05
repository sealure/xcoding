package executor

import (
	"context"
	"sync"
)

type LogLine struct {
	Seq     uint64
	Content string
}

type LogWriter interface {
	WriteLogs(ctx context.Context, buildID uint64, lines []*LogLine) error
}

// LogBatcher 简单批处理器：累积若干行后调用批量写接口（并发安全）
type LogBatcher struct {
	mu       sync.Mutex
	writer   LogWriter
	buildID  uint64
	lines    []*LogLine
	capacity int
}

func NewLogBatcher(writer LogWriter, buildID uint64, capacity int) *LogBatcher {
	if capacity <= 0 {
		capacity = 50
	}
	return &LogBatcher{writer: writer, buildID: buildID, capacity: capacity}
}

// Append 追加单行日志到缓冲；达到容量后自动 Flush
func (b *LogBatcher) Append(ctx context.Context, seq uint64, content string) {
	b.mu.Lock()
	b.lines = append(b.lines, &LogLine{Seq: seq, Content: content})
	needFlush := len(b.lines) >= b.capacity
	b.mu.Unlock()
	if needFlush {
		_ = b.Flush(ctx)
	}
}

// Flush 将缓冲区的日志批量写入后端（若 writer 存在）
func (b *LogBatcher) Flush(ctx context.Context) error {
	b.mu.Lock()
	if len(b.lines) == 0 || b.writer == nil {
		b.mu.Unlock()
		return nil
	}
	lines := b.lines
	b.lines = nil
	b.mu.Unlock()
	return b.writer.WriteLogs(ctx, b.buildID, lines)
}
