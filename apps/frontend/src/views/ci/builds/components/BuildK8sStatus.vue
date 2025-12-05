<template>
  <div class="k8s-status">
    <div class="k8s-actions">
      <el-switch v-model="k8sOnlyAbnormal" active-text="只看异常" inactive-text="显示全部" />
      <el-button size="small" @click="fetchK8sStatus">刷新</el-button>
    </div>

    <div class="cards">
      <div v-for="job in k8sJobs" :key="job.job_name" class="card">
        <div class="card-header">
          <div class="card-title">
            <el-tag size="small" type="info">Job</el-tag>
            <span class="name">{{ job.job_name }}</span>
          </div>
          <div class="card-meta">
            <span class="count">Pods: {{ (job.pods || []).length }}</span>
          </div>
        </div>
        <div class="card-body">
          <div class="section">
            <div class="section-title">条件</div>
            <div class="cond-list">
              <div v-for="c in job.conditions" :key="c.type + c.reason" class="cond-item">
                <el-tag size="small" :type="k8sCondType(c)">{{ c.type }}</el-tag>
                <span class="reason" v-if="c.reason">{{ c.reason }}</span>
              </div>
              <div v-if="!(job.conditions && job.conditions.length)" class="empty">—</div>
            </div>
          </div>
          <div class="section">
            <div class="section-title">Pods</div>
            <div class="pods-list">
              <div v-for="p in job.pods" :key="p.name" class="pod-item">
                <el-tag size="small" :type="podType(p)">{{ p.name }}</el-tag>
                <span class="phase">{{ p.phase }}</span>
                <span class="node">{{ p.node || '—' }}</span>
                <el-tag v-if="p.reason" size="small" type="danger" class="reason-tag">{{ p.reason }}</el-tag>
              </div>
              <div v-if="!(job.pods && job.pods.length)" class="empty">—</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="pagination-container">
      <el-pagination
        v-model:current-page="k8sPage"
        v-model:page-size="k8sPageSize"
        :page-sizes="[10, 20, 50]"
        :total="k8sTotal"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleK8sSizeChange"
        @current-change="handleK8sPageChange"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'
import { getExecutorK8sStatus } from '@/api/ci/builds'
const props = defineProps({ buildId: { type: [String, Number], required: true } })
const k8sJobs = ref([])
const k8sPage = ref(1)
const k8sPageSize = ref(10)
const k8sTotal = ref(0)
const k8sOnlyAbnormal = ref(false)
const k8sCondType = (c) => {
  if (!c) return ''
  const t = String(c.type || '').toLowerCase()
  if (t.includes('failed')) return 'danger'
  if (t.includes('complete') || t.includes('succeeded')) return 'success'
  return 'info'
}
const fetchK8sStatus = async () => {
  try {
    const res = await getExecutorK8sStatus(String(props.buildId), '', k8sPage.value, k8sPageSize.value)
    const all = res?.jobs || res?.data || []
    const abnormal = (all || []).filter(j => {
      const hasUnsched = (j.conditions || []).some(c => String(c.reason).toLowerCase() === 'unschedulable')
      const hasFailed = (j.failed && j.failed > 0) || (j.conditions || []).some(c => String(c.type).toLowerCase().includes('failed'))
      const podsFailed = (j.pods || []).some(p => String(p.phase).toLowerCase() === 'failed')
      return hasUnsched || hasFailed || podsFailed
    })
    const data = k8sOnlyAbnormal.value ? abnormal : all
    k8sJobs.value = data
    const totalItems = (k8sOnlyAbnormal.value ? abnormal.length : (res?.pagination?.total_items || (all.length || 0)))
    k8sTotal.value = totalItems
  } catch (e) { console.error('获取K8s状态失败:', e) }
}
const handleK8sSizeChange = (s) => { k8sPageSize.value = s; fetchK8sStatus() }
const handleK8sPageChange = (p) => { k8sPage.value = p; fetchK8sStatus() }
const podType = (p) => {
  const phase = String(p?.phase || '').toLowerCase()
  if (phase.includes('failed')) return 'danger'
  if (phase.includes('running')) return 'success'
  if (phase.includes('pending')) return 'warning'
  return 'info'
}
watch(() => props.buildId, async () => { k8sPage.value = 1; await nextTick(); fetchK8sStatus() }, { immediate: true })
</script>

<style scoped>
.k8s-status { display: flex; flex-direction: column; gap: 12px; flex: 1 1 auto; min-height: 0; height: 100%; }
.k8s-actions { display: flex; gap: 10px; align-items: center; }
.cards { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; flex: 1 1 auto; min-height: 0; overflow: auto; }
.card { border: 1px solid #ebeef5; border-radius: 8px; background: #fff; padding: 10px 12px; display: flex; flex-direction: column; }
.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.card-title { display: flex; align-items: center; gap: 8px; font-weight: 600; }
.card-title .name { color: #303133; }
.card-meta .count { color: #909399; font-size: 12px; }
.section { display: flex; flex-direction: column; gap: 6px; }
.section + .section { margin-top: 6px; }
.section-title { color: #606266; font-size: 13px; }
.cond-list, .pods-list { display: flex; flex-direction: column; gap: 6px; }
.cond-item, .pod-item { display: flex; align-items: center; gap: 8px; }
.pod-item .phase { color: #606266; }
.pod-item .node { color: #909399; font-size: 12px; }
.reason-tag { margin-left: auto; }
.empty { color: #c0c4cc; font-size: 13px; }
.pagination-container { display: flex; justify-content: flex-end; }
@media (max-width: 1200px) { .cards { grid-template-columns: 1fr; } }
</style>
