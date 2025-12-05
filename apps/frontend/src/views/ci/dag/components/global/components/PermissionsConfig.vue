<template>
  <el-form label-width="120px">
    <el-form-item label="permissions（键值对）">
      <div>
        <div v-for="(pair, idx) in localPairs" :key="`perm-${idx}`" class="env-pair-row">
          <el-input v-model="pair.key" placeholder="SCOPE" style="width: 40%; margin-right: 8px" />
          <el-input v-model="pair.value" placeholder="read/write/none" style="width: 48%; margin-right: 8px" />
          <el-button type="danger" plain size="small" @click="removePair(idx)">删除</el-button>
        </div>
        <el-button type="primary" plain size="small" @click="addPair">新增权限项</el-button>
      </div>
    </el-form-item>
  </el-form>
</template>

<script setup>
import { computed } from 'vue'
const props = defineProps({
  pairs: { type: Array, default: () => [] }
})
const emit = defineEmits(['update:pairs'])
const localPairs = computed({
  get: () => props.pairs,
  set: (v) => emit('update:pairs', v)
})
const addPair = () => { localPairs.value = [...localPairs.value, { key: '', value: '' }] }
const removePair = (idx) => { localPairs.value = localPairs.value.filter((_, i) => i !== idx) }
</script>

<style scoped>
.env-pair-row { margin-bottom: 8px; display: flex; align-items: center; }
</style>