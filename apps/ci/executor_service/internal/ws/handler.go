package ws

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"xcoding/apps/ci/executor_service/internal/executor"
	"xcoding/apps/ci/executor_service/models"
	civ1 "xcoding/gen/go/ci/v1"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	db  *gorm.DB
	k8s *executor.K8sEnv
}

func (h *Handler) isTerminalStatus(s civ1.BuildStatus) bool {
	return s == civ1.BuildStatus_BUILD_STATUS_SUCCEEDED || s == civ1.BuildStatus_BUILD_STATUS_FAILED || s == civ1.BuildStatus_BUILD_STATUS_CANCELLED
}

func (h *Handler) sendLogDelta(conn *websocket.Conn, buildID uint64, afterID uint64) (uint64, error) {
	type LogResult struct {
		ID           uint64
		Content      string
		JobID        uint64
		JobName      string
		StepID       uint64
		StepName     string
		CreatedAt    time.Time
		JobStartedAt *time.Time
		JobStatus    string
		StepStatus   string
	}
	var logs []LogResult
	err := h.db.Model(&models.BuildStepLogChunk{}).
		Select("build_step_log_chunks.id AS id, build_step_log_chunks.content AS content, build_step_log_chunks.created_at AS created_at, s.id AS step_id, s.name AS step_name, s.status AS step_status, s.job_name AS job_name, j.id AS job_id, j.status AS job_status, j.started_at AS job_started_at").
		Joins("JOIN build_steps s ON s.id = build_step_log_chunks.build_step_id").
		Joins("JOIN build_jobs j ON j.build_id = s.build_id AND j.name = s.job_name").
		Where("s.build_id = ? AND build_step_log_chunks.id > ?", buildID, afterID).
		Order("build_step_log_chunks.id ASC").
		Limit(100).
		Find(&logs).Error
	if err != nil {
		return afterID, nil
	}
	if len(logs) == 0 {
		return afterID, nil
	}
	jobs := make(map[uint64]map[string]any)
	jobSteps := make(map[uint64]map[uint64]map[string]any)
	nextID := afterID
	for _, l := range logs {
		if strings.TrimSpace(l.Content) == "" {
			if l.ID > nextID {
				nextID = l.ID
			}
			continue
		}
		j := jobs[l.JobID]
		if j == nil {
			ct := l.CreatedAt
			if l.JobStartedAt != nil {
				ct = *l.JobStartedAt
			}
			j = map[string]any{
				"id":         l.JobID,
				"name":       l.JobName,
				"status":     l.JobStatus,
				"created_at": ct,
				"step":       []map[string]any{},
			}
			jobs[l.JobID] = j
			jobSteps[l.JobID] = make(map[uint64]map[string]any)
		}
		stMap := jobSteps[l.JobID]
		st := stMap[l.StepID]
		if st == nil {
			st = map[string]any{"id": l.StepID, "name": l.StepName, "status": l.StepStatus, "logs": []map[string]any{}}
			stMap[l.StepID] = st
			j["step"] = append(j["step"].([]map[string]any), st)
		}
		ls := strings.Split(l.Content, "\n")
		for i, ln := range ls {
			s := strings.TrimSpace(ln)
			if s == "" {
				continue
			}
			st["logs"] = append(st["logs"].([]map[string]any), map[string]any{"id": l.ID, "seq": i, "content": s, "created_at": l.CreatedAt})
		}
		if l.ID > nextID {
			nextID = l.ID
		}
	}
	if len(jobs) == 0 {
		return nextID, nil
	}
	var jobsArr []map[string]any
	for _, v := range jobs {
		jobsArr = append(jobsArr, v)
	}
	msg := map[string]any{"type": "log", "data": map[string]any{"jobs": jobsArr}}
	if err := conn.WriteJSON(msg); err != nil {
		return nextID, err
	}
	return nextID, nil
}

func (h *Handler) sendStatus(conn *websocket.Conn, b models.Build) error {
	type BuildWithStrStatus struct {
		models.Build
		Status string `json:"status"`
	}
	msg := map[string]any{
		"type": "status",
		"data": BuildWithStrStatus{Build: b, Status: civ1.BuildStatus(b.Status).String()},
	}
	return conn.WriteJSON(msg)
}

func (h *Handler) sendK8sStatus(conn *websocket.Conn, buildID uint64) error {
	if h.k8s == nil {
		return nil
	}
	k8sStatus, err := h.k8s.ListBuildJobPodStatus(context.Background(), buildID)
	if err != nil {
		return nil
	}
	msg := map[string]any{"type": "k8s_status", "data": k8sStatus}
	return conn.WriteJSON(msg)
}

func (h *Handler) sendDAG(conn *websocket.Conn, buildID uint64) error {
	var jobs []models.BuildJob
	if err := h.db.Where("build_id = ?", buildID).Order("index asc").Find(&jobs).Error; err != nil || len(jobs) == 0 {
		return nil
	}
	var steps []models.BuildStep
	h.db.Where("build_id = ?", buildID).Order("job_name asc, index asc").Find(&steps)
	var edges []models.BuildJobEdge
	h.db.Where("build_id = ?", buildID).Find(&edges)
	msg := map[string]any{
		"type": "dag",
		"data": map[string]any{"jobs": jobs, "steps": steps, "edges": edges},
	}
	return conn.WriteJSON(msg)
}

// sendBuildStatus 组合发送构建整体状态（构建信息、DAG边、Job/Step 状态以及增量日志）
// - afterID 为增量日志的游标（基于 build_step_log_chunks.id）
// - 返回最新的游标 ID
func (h *Handler) sendBuildStatus(conn *websocket.Conn, buildID uint64, afterID uint64) (uint64, error) {
	var build models.Build
	if err := h.db.First(&build, buildID).Error; err != nil {
		return afterID, err
	}

	var edges []models.BuildJobEdge
	h.db.Where("build_id = ?", buildID).Find(&edges)

	var jobRows []models.BuildJob
	h.db.Where("build_id = ?", buildID).Order("index asc").Find(&jobRows)

	var stepRows []models.BuildStep
	h.db.Where("build_id = ?", buildID).Order("job_name asc, index asc").Find(&stepRows)

	jobs := make(map[string]map[string]any)
	stepMap := make(map[uint64]map[string]any)

	for _, j := range jobRows {
		ct := build.CreatedAt
		if j.StartedAt != nil {
			ct = *j.StartedAt
		}
		jobs[j.Name] = map[string]any{
			"id":         j.ID,
			"name":       j.Name,
			"status":     j.Status,
			"created_at": ct,
			"step":       []map[string]any{},
		}
	}

	for _, st := range stepRows {
		j := jobs[st.JobName]
		if j == nil {
			// 步骤所属 Job 尚未出现（理论上不会发生），跳过
			continue
		}
		sm := map[string]any{"id": st.ID, "name": st.Name, "status": st.Status, "logs": []map[string]any{}}
		j["step"] = append(j["step"].([]map[string]any), sm)
		stepMap[st.ID] = sm
	}

	// 追加增量日志
	var logRows []models.BuildStepLogChunk
	var stepIDs []uint64
	stepIDs = make([]uint64, 0, len(stepMap))
	for sid := range stepMap {
		stepIDs = append(stepIDs, sid)
	}
	nextID := afterID
	if len(stepIDs) > 0 {
		if err := h.db.Where("build_step_id IN ? AND id > ?", stepIDs, afterID).Order("id ASC").Limit(500).Find(&logRows).Error; err != nil {
			// 不因日志查询失败中断，保留游标
			log.Printf("sendBuildStatus: log query error: %v", err)
		} else {
			for _, l := range logRows {
				sm := stepMap[l.BuildStepID]
				if sm == nil {
					continue
				}
				s := strings.TrimSpace(l.Content)
				if s == "" {
					if l.ID > nextID {
						nextID = l.ID
					}
					continue
				}
				sm["logs"] = append(sm["logs"].([]map[string]any), map[string]any{"content": s, "created_at": l.CreatedAt})
				if l.ID > nextID {
					nextID = l.ID
				}
			}
		}
	}

	// 构造 jobs 数组（按原始 jobRows 顺序）
	var jobsArr []map[string]any
	for _, jr := range jobRows {
		if jm := jobs[jr.Name]; jm != nil {
			jobsArr = append(jobsArr, jm)
		}
	}

	type BuildWithStrStatus struct {
		models.Build
		Status string `json:"status"`
	}
	payload := map[string]any{
		"type": "build_status",
		"data": map[string]any{
			"build": BuildWithStrStatus{Build: build, Status: civ1.BuildStatus(build.Status).String()},
			"edges": edges,
			"jobs":  jobsArr,
		},
	}
	if err := conn.WriteJSON(payload); err != nil {
		return nextID, err
	}
	return nextID, nil
}

func NewHandler(db *gorm.DB) *Handler {
	k8s, _ := executor.NewK8sEnv()
	return &Handler{db: db, k8s: k8s}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Log all headers for debugging
	for k, v := range r.Header {
		log.Printf("Header %s: %v", k, v)
	}

	// Path: /ci_service/api/v1/executor/ws/builds/{id}
	vars := strings.Split(r.URL.Path, "/")
	if len(vars) < 2 {
		http.Error(w, "invalid build id", http.StatusBadRequest)
		return
	}
	buildIDStr := vars[len(vars)-1]
	// Handle query params if needed, or just parse ID
	// The path is .../builds/{id}

	buildID, err := strconv.ParseUint(buildIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid build id", http.StatusBadRequest)
		return
	}

	offsetStr := r.URL.Query().Get("offset")
	var offset uint64 = 0
	if offsetStr != "" {
		offset, _ = strconv.ParseUint(offsetStr, 10, 64)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WS connected: %d", buildID)
	h.streamBuildData(conn, buildID, offset)
}

func (h *Handler) streamBuildData(conn *websocket.Conn, buildID uint64, lastLogSeq uint64) {

	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	// 使用日志 ID 做增量游标
	lastLogID := lastLogSeq

	for {
		select {
		case <-ticker.C:
			var err error
			lastLogID, err = h.sendBuildStatus(conn, buildID, lastLogID)
			if err != nil {
				return
			}
		}
	}
}
