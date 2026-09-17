<template>
  <div>
    <div class="toolbar">
      <div>
        <h2 class="page-title">测年送检</h2>
        <p class="page-sub">围绕出土文物发起碳十四 / 热释光测年送检，跟踪送检与结果回填</p>
      </div>
      <button class="btn" @click="openCreate()">新建送检单</button>
    </div>

    <div class="card">
      <div class="filters">
        <label>
          状态筛选
          <select v-model="filterStatus" @change="load">
            <option value="">全部状态</option>
            <option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
          </select>
        </label>
        <label>
          测年方法
          <select v-model="filterMethod" @change="load">
            <option value="">全部方法</option>
            <option value="c14">c14 碳十四</option>
            <option value="tl">tl 热释光</option>
          </select>
        </label>
      </div>

      <table class="table">
        <thead>
          <tr>
            <th>#</th>
            <th>送检文物</th>
            <th>实验室</th>
            <th>方法</th>
            <th>状态</th>
            <th>送检时间</th>
            <th>结果时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in list" :key="item.id">
            <td>{{ item.id }}</td>
            <td>
              <router-link v-if="item.linkedFind" class="find-link" :to="{ name: 'finds' }">
                {{ item.linkedFind.registerNo }}
              </router-link>
              <span v-else class="muted">文物缺失</span>
              <div class="muted small">
                {{ item.linkedFind?.artifactType }} ·
                {{ item.linkedFind?.unit?.site?.name || '' }} / {{ item.linkedFind?.unit?.code || '-' }}
              </div>
            </td>
            <td>{{ item.labName }}</td>
            <td><span class="tag">{{ methodLabel(item.method) }}</span></td>
            <td><span class="status" :class="'st-' + item.status">{{ statusLabel(item.status) }}</span></td>
            <td>{{ formatDT(item.submittedAt) }}</td>
            <td>{{ formatDT(item.resultedAt) }}</td>
            <td>
              <button class="btn secondary small" @click="openDetail(item)">详情</button>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-if="!list.length" class="page-sub">暂无数据</p>
      <p v-if="error" class="error">{{ error }}</p>
    </div>

    <!-- 新建 / 编辑（仅 draft） -->
    <div v-if="showForm" class="modal-mask" @click.self="showForm = false">
      <div class="modal">
        <h3>{{ editing ? '编辑送检单（草稿）' : '新建送检单' }}</h3>
        <div class="form-grid">
          <label class="full">
            送检文物（必选且仅选一件）
            <select v-model.number="form.linkedFindId" :disabled="editing === false && presetFindId">
              <option :value="0" disabled>请选择出土文物</option>
              <option v-for="f in finds" :key="f.id" :value="f.id">
                {{ f.registerNo }}｜{{ f.artifactType }}｜{{ f.unit?.site?.name || '' }} / {{ f.unit?.code || '-' }}
              </option>
            </select>
          </label>
          <label>
            承测实验室
            <input v-model="form.labName" placeholder="如：社科院考古所碳十四实验室" />
          </label>
          <label>
            测年方法
            <select v-model="form.method">
              <option value="c14">c14 碳十四</option>
              <option value="tl">tl 热释光</option>
            </select>
          </label>
        </div>
        <p class="page-sub hint">送检对象必须且只能关联一件出土文物；系统暂无样品（Sample）表，暂不支持挂样品。</p>
        <p v-if="formError" class="error">{{ formError }}</p>
        <div class="modal-actions">
          <button class="btn secondary" @click="showForm = false">取消</button>
          <button class="btn" @click="save">保存草稿</button>
        </div>
      </div>
    </div>

    <!-- 详情 / 流转 -->
    <div v-if="showDetail" class="modal-mask" @click.self="showDetail = false">
      <div class="modal modal-lg">
        <h3>送检单 #{{ detail.id }}</h3>
        <div class="detail-grid">
          <div><span class="muted">状态</span>
            <span class="status" :class="'st-' + detail.status">{{ statusLabel(detail.status) }}</span>
          </div>
          <div><span class="muted">测年方法</span>{{ methodLabel(detail.method) }}</div>
          <div class="full"><span class="muted">承测实验室</span>{{ detail.labName }}</div>
          <div class="full">
            <span class="muted">送检文物</span>
            <template v-if="detail.linkedFind">
              {{ detail.linkedFind.registerNo }}｜{{ detail.linkedFind.artifactType }}｜
              {{ detail.linkedFind.unit?.site?.name || '' }} / {{ detail.linkedFind.unit?.code || '-' }}
              <div class="muted small">{{ detail.linkedFind.description }}</div>
            </template>
            <span v-else class="muted">文物缺失</span>
          </div>
          <div><span class="muted">送检时间</span>{{ formatDT(detail.submittedAt) }}</div>
          <div><span class="muted">结果时间</span>{{ formatDT(detail.resultedAt) }}</div>
          <div class="full">
            <span class="muted">测年结果</span>
            <div class="result-box" v-if="detail.resultText">{{ detail.resultText }}</div>
            <span v-else class="muted">尚未回填</span>
          </div>
        </div>

        <p v-if="detailError" class="error">{{ detailError }}</p>

        <div class="modal-actions wrap">
          <template v-if="detail.status === 'draft'">
            <button class="btn danger" @click="remove(detail)">删除草稿</button>
            <button class="btn secondary" @click="openEditFromDetail">编辑</button>
            <button class="btn" @click="doTransition('submit')">提交送检</button>
          </template>
          <template v-else-if="detail.status === 'submitted'">
            <button class="btn danger" @click="doTransition('void')">作废</button>
            <button class="btn" @click="openResult">回填结果</button>
          </template>
          <template v-else>
            <span class="muted terminal-hint">
              {{ detail.status === 'resulted' ? '已出结果' : '已作废' }}，终态不可再流转或修改关联。
            </span>
            <button class="btn secondary" @click="showDetail = false">关闭</button>
          </template>
        </div>
      </div>
    </div>

    <!-- 回填结果 -->
    <div v-if="showResultBox" class="modal-mask" @click.self="showResultBox = false">
      <div class="modal">
        <h3>回填测年结果</h3>
        <label class="block-label">
          结果内容（可留空后补录，但一旦出结果即进入终态）
          <textarea v-model="resultText" placeholder="如：碳十四测年 BP 3200±30，校正后约公元前 1510–1450 年"></textarea>
        </label>
        <p v-if="detailError" class="error">{{ detailError }}</p>
        <div class="modal-actions">
          <button class="btn secondary" @click="showResultBox = false">取消</button>
          <button class="btn" @click="doTransition('result')">确认出结果</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api/http'

const route = useRoute()
const router = useRouter()

const statusOptions = [
  { value: 'draft', label: '草稿' },
  { value: 'submitted', label: '已送检' },
  { value: 'resulted', label: '已出结果' },
  { value: 'void', label: '已作废' }
]

const list = ref([])
const finds = ref([])
const filterStatus = ref('')
const filterMethod = ref('')
const error = ref('')

const showForm = ref(false)
const formError = ref('')
const editing = ref(false)
const presetFindId = ref(0)
const form = reactive({ id: null, linkedFindId: 0, labName: '', method: 'c14' })

const showDetail = ref(false)
const detailError = ref('')
const detail = ref({})

const showResultBox = ref(false)
const resultText = ref('')

function statusLabel(s) {
  return statusOptions.find((x) => x.value === s)?.label || s
}
function methodLabel(m) {
  return m === 'tl' ? 'tl 热释光' : 'c14 碳十四'
}
function formatDT(v) {
  if (!v) return '-'
  return String(v).replace('T', ' ').slice(0, 16)
}

async function loadFinds() {
  const { data } = await api.get('/finds')
  finds.value = data
}

async function load() {
  error.value = ''
  try {
    const params = {}
    if (filterStatus.value) params.status = filterStatus.value
    if (filterMethod.value) params.method = filterMethod.value
    const { data } = await api.get('/dating-submissions', { params })
    list.value = data
  } catch (e) {
    error.value = e.response?.data?.error || '加载失败'
  }
}

function resetForm(findId) {
  Object.assign(form, {
    id: null,
    linkedFindId: findId || 0,
    labName: '',
    method: 'c14'
  })
  presetFindId.value = findId || 0
  editing.value = false
  formError.value = ''
}

// 从“出土文物”页跳转：/dating-submissions?findId=5&new=1
function openCreate(findId) {
  resetForm(findId)
  showForm.value = true
}

function openEditFromDetail() {
  const d = detail.value
  Object.assign(form, {
    id: d.id,
    linkedFindId: d.linkedFindId || 0,
    labName: d.labName,
    method: d.method
  })
  presetFindId.value = 0
  editing.value = true
  formError.value = ''
  showDetail.value = false
  showForm.value = true
}

async function save() {
  formError.value = ''
  if (!form.linkedFindId) {
    formError.value = '请选择一件送检文物'
    return
  }
  if (!form.labName.trim()) {
    formError.value = '请填写承测实验室'
    return
  }
  const payload = {
    labName: form.labName.trim(),
    method: form.method,
    linkedFindId: form.linkedFindId,
    linkedSampleId: null
  }
  try {
    if (editing.value) {
      await api.put(`/dating-submissions/${form.id}`, payload)
    } else {
      await api.post('/dating-submissions', payload)
    }
    showForm.value = false
    router.replace({ name: 'dating' })
    await load()
  } catch (e) {
    formError.value = e.response?.data?.error || '保存失败'
  }
}

async function openDetail(item) {
  detailError.value = ''
  try {
    const { data } = await api.get(`/dating-submissions/${item.id}`)
    detail.value = data
    showDetail.value = true
  } catch (e) {
    detailError.value = e.response?.data?.error || '加载详情失败'
  }
}

async function remove(item) {
  if (!confirm(`确认删除送检单 #${item.id}（草稿）？`)) return
  try {
    await api.delete(`/dating-submissions/${item.id}`)
    showDetail.value = false
    await load()
  } catch (e) {
    detailError.value = e.response?.data?.error || '删除失败'
  }
}

function openResult() {
  resultText.value = detail.value.resultText || ''
  detailError.value = ''
  showResultBox.value = true
}

async function doTransition(action) {
  if (action === 'submit' && !confirm('确认提交送检？提交后在结果回填前不可修改关联。')) return
  if (action === 'void' && !confirm('确认作废该送检单？作废为终态操作，不可撤销。')) return
  detailError.value = ''
  try {
    const { data } = await api.post(`/dating-submissions/${detail.value.id}/transition`, {
      action,
      resultText: action === 'result' ? resultText.value : undefined
    })
    showResultBox.value = false
    detail.value = data
    await load()
    if (action === 'result') showDetail.value = true
  } catch (e) {
    detailError.value = e.response?.data?.error || '流转失败'
  }
}

onMounted(async () => {
  await loadFinds()
  await load()
  if (route.query.new === '1' && route.query.findId) {
    openCreate(Number(route.query.findId))
  }
})
</script>

<style scoped>
.filters {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.filters label {
  min-width: 200px;
}

.find-link {
  color: var(--accent);
  font-weight: 600;
}

.muted {
  color: var(--muted);
}

.small {
  font-size: 0.8rem;
}

.status {
  display: inline-block;
  padding: 0.15rem 0.6rem;
  border-radius: 999px;
  font-size: 0.8rem;
  font-weight: 600;
}

.st-draft {
  background: #ece4d6;
  color: #7a6648;
}

.st-submitted {
  background: #dde7f0;
  color: #2f5a7c;
}

.st-resulted {
  background: #dcece2;
  color: var(--ok);
}

.st-void {
  background: #efd9d4;
  color: var(--danger);
}

.modal-lg {
  width: min(720px, 100%);
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem 1.2rem;
}

.detail-grid .full {
  grid-column: 1 / -1;
}

.detail-grid span.muted {
  display: block;
  font-size: 0.82rem;
  margin-bottom: 0.2rem;
}

.result-box {
  background: #f6f0e5;
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 0.7rem 0.85rem;
  white-space: pre-wrap;
}

.hint {
  margin-top: 0.7rem;
}

.block-label {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.wrap {
  flex-wrap: wrap;
}

.terminal-hint {
  margin-right: auto;
  align-self: center;
}
</style>
