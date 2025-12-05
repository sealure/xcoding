<template>
  <div class="job-container" v-if="job">
    <div class="job-header">
      <span class="job-title">{{ jobTitle }}</span>
      <span class="job-if" v-if="jobIf">{{ jobIf }}</span>
      <span class="job-matrix" v-if="matrixText">{{ matrixText }}</span>
      <el-button size="small" @click="$emit('toggle-collapse', jid)">
        {{ collapsed ? '展开' : '折叠' }}
      </el-button>
      <el-button size="small" type="primary" @click="$emit('show-job', jid)">配置</el-button>
    </div>
    <!-- 预留：将来可以在此渲染步骤静态视图或占位 -->
  </div>
</template>

<script>
import { matrixHint, jobIfType } from '../../utils'

export default {
  name: 'JobContainer',
  props: {
    jid: { type: String, default: '' },
    job: { type: Object, default: () => ({}) },
    collapsed: { type: Boolean, default: false },
  },
  emits: ['toggle-collapse', 'show-job'],
  computed: {
    jobTitle() {
      try { return this.job?.name || this.jid || '' } catch (_) { return this.jid || '' }
    },
    jobIf() {
      try { return jobIfType(this.job) } catch (_) { return '' }
    },
    matrixText() {
      try { const t = matrixHint(this.job); return t ? `[${t}]` : '' } catch (_) { return '' }
    },
  },
}
</script>

<style scoped>
.job-container { display: flex; flex-direction: column; gap: 8px; }
.job-header { display: flex; align-items: center; gap: 8px; }
.job-title { font-weight: 600; color: #606266; }
.job-if { font-size: 12px; color: #909399; }
.job-matrix { font-size: 12px; color: #909399; }
</style>
