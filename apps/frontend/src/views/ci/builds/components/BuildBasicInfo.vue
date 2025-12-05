<template>
  <div class="basic-info-wrap">
    <el-descriptions title="基本信息" :column="3" border>
      <el-descriptions-item label="状态">
        <el-tag :type="statusTagType(detail?.status)" effect="light" round>{{ formatStatus(detail?.status) }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="流水线ID">{{ detail?.pipeline_id || '—' }}</el-descriptions-item>
      <el-descriptions-item label="触发者">{{ detail?.triggered_by || '—' }}</el-descriptions-item>
      <el-descriptions-item label="分支">{{ detail?.branch || '—' }}</el-descriptions-item>
      <el-descriptions-item label="提交">{{ detail?.commit_sha || '—' }}</el-descriptions-item>
      <el-descriptions-item label="创建时间">{{ formatDate(detail?.created_at) }}</el-descriptions-item>
      <el-descriptions-item label="开始时间">{{ formatDate(detail?.started_at) }}</el-descriptions-item>
      <el-descriptions-item label="结束时间">{{ formatDate(detail?.finished_at) }}</el-descriptions-item>
    </el-descriptions>
  </div>
</template>

<script setup>
import { computed } from 'vue'
const props = defineProps({ detail: { type: Object, default: () => ({}) } })
const detail = computed(() => props.detail || {})
const formatStatus = (val) => {
  const map = {
    BUILD_STATUS_PENDING: '待下发',
    BUILD_STATUS_QUEUED: '队列中',
    BUILD_STATUS_RUNNING: '运行中',
    BUILD_STATUS_SUCCEEDED: '成功',
    BUILD_STATUS_FAILED: '失败',
    BUILD_STATUS_CANCELLED: '已取消'
  }
  return map[val] || val || '—'
}
const formatDate = (ts) => { try { return ts ? new Date(ts).toLocaleString('zh-CN') : '—' } catch { return '—' } }
const statusTagType = (val) => {
  const map = {
    BUILD_STATUS_PENDING: 'info',
    BUILD_STATUS_QUEUED: 'warning',
    BUILD_STATUS_RUNNING: 'primary',
    BUILD_STATUS_SUCCEEDED: 'success',
    BUILD_STATUS_FAILED: 'danger',
    BUILD_STATUS_CANCELLED: 'info'
  }
  return map[val] || 'info'
}
</script>

<style scoped>
</style>
<style scoped>
.basic-info-wrap { width: 100%; }
.basic-info-wrap :deep(.el-descriptions) { width: 100%; }
</style>
