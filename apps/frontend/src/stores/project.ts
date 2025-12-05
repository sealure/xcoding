import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getProjectList } from '@/api/project'

type ProjectOption = { id: string | number; name?: string; [k: string]: any }
type ProjectListResponse = { data?: ProjectOption[] }

export const useProjectStore = defineStore('project', () => {
  const selectedProject = ref<ProjectOption | null>(null)
  const projectOptions = ref<ProjectOption[]>([])

  const setSelectedProject = (project: ProjectOption | null) => {
    selectedProject.value = project
    if (project?.id) localStorage.setItem('selected_project_id', String(project.id))
    else localStorage.removeItem('selected_project_id')
  }

  const loadPersisted = async () => {
    const id = localStorage.getItem('selected_project_id')
    if (!id) return
    // 尝试在当前 options 中找到并恢复
    if (projectOptions.value?.length) {
      const p = projectOptions.value.find((x) => String(x.id) === String(id))
      if (p) selectedProject.value = p
      return
    }
    // 否则请求一次列表进行恢复
    try {
      const res = await getProjectList({ page: 1, page_size: 100 }) as ProjectListResponse
      projectOptions.value = res.data || []
      const p = projectOptions.value.find((x) => String(x.id) === String(id))
      if (p) selectedProject.value = p
    } catch (_) {}
  }

  const fetchProjectOptions = async () => {
    const res = await getProjectList({ page: 1, page_size: 100 }) as ProjectListResponse
    projectOptions.value = res.data || []
    // 若无已选，默认不自动选中，等待用户弹窗选择
  }

  return {
    selectedProject,
    projectOptions,
    setSelectedProject,
    fetchProjectOptions,
    loadPersisted,
  }
})
