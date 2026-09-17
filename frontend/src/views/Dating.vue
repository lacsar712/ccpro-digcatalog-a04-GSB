<template>
  <div>
    <div class="toolbar">
      <div>
        <h2 class="page-title">测年送检</h2>
        <p class="page-sub">文物碳十四 / 热释光测年送样、流转与结果登记（每单必须关联一件出土文物）</p>
      </div>
      <button class="btn" @click="openCreate">新建送检单</button>
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
            <option value="c14">碳十四（C14）</option>
            <option value="tl">热释光（TL）</option>
          </select>
        </label>
      </div>

      <table class="table">
        <thead>
          <tr>
            <th>#</th>
            <th>承测实验室</th>
            <th>方法</th>
            <th>关联文物</th>
            <th>工地 / 探方</th>
            <th>状态</th>
            <th>送检时间</th>
            <th>出结果时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in list" :key="item.id">
            <td>{{ item.id }}</td>
            <td>{{ item.labName }}</td>
            <td>{{ methodLabel(item.method) }}</td>
            <td>
              <router-link :to="{ name: 'finds' }" class="find-link">
                {{ item.linkedFind?.registerNo || '-' }}
              </router-link>
              <span class="muted">（{{ item.linkedFind?.artifactType || '-' }}）</span>
            </td>
            <td>{{ findLocation(item.linkedFind) }}</td>
            <td><span class="status-tag" :class="'st-' + item.status">{{ statusLabel(item.status) }}</span></td>
            <td>{{ formatDate(item.submittedAt) }}</td>
            <td>{{ formatDate(item.resultedAt) }}</td>
            <td>
              <button class="btn secondary small" @click="openDetail(item)">详情</button>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-if="!list.length" class="page-sub">暂无送检单</p>
      <p v-if="error" class="error">{{ error }}</p>
    </div>

    <!-- 创建 / 编辑草稿 -->
    <div v-if="showEditor" class="modal-mask" @click.self="showEditor = false">
      <div class="modal">
        <h3>{{ form.id ? '编辑送检单（草稿）' : '新建送检单' }}</h3>
        <div class="form-grid">
          <label class="full">
            关联文物 <span class="required">*</span>
            <select v-model.number="form.linkedFindId">
              <option :value="0" disabled>请选择出土文物</option>
              <option v-for="f in finds" :key="f.id" :value="f.id">
                {{ f.registerNo }} · {{ f.artifactType }} · {{ f.unit?.site?.name || '' }} / {{ f.unit?.code || '' }}
              </option>
            </select>
          </label>
          <label>
            承测实验室 <span class="required">*</span>
            <input v-model="form.labName" placeholder="如：社科院考古所碳十四实验室" />
          </label>
          <label>
            测年方法 <span class="required">*</span>
            <select v-model="form.method">
              <option value="c14">碳十四（C14）</option>
              <option value="tl">热释光（TL）</option>
            </select>
          </label>
        </div>
        <p class="editor-hint">新建送检单默认为「草稿」状态；送检后文物关联不可再更改。</p>
        <p v-if="formError" class="error">{{ formError }}</p>
        <div class="modal-actions">
          <button class="btn secondary" @click="showEditor = false">取消</button>
          <button class="btn" @click="save">保存草稿</button>
        </div>
      </div>
    </div>

    <!-- 详情 / 流转 -->
    <div v-if="detail" class="modal-mask" @click.self="detail = null">
      <div class="modal modal-wide">
        <h3>送检单 #{{ detail.id }} <span class="status-tag" :class="'st-' + detail.status">{{ statusLabel(detail.status) }}</span></h3>

        <div class="detail-grid">
          <div>
            <div class="detail-label">承测实验室</div>
            <div>{{ detail.labName }}</div>
          </div>
          <div>
            <div class="detail-label">测年方法</div>
            <div>{{ methodLabel(detail.method) }}</div>
          </div>
          <div>
            <div class="detail-label">关联文物</div>
            <div v-if="detail.linkedFind">
              {{ detail.linkedFind.registerNo }} · {{ detail.linkedFind.artifactType }}
              <span class="muted">（{{ findLocation(detail.linkedFind) }}）</span>
            </div>
            <div v-else>-</div>
          </div>
          <div>
            <div class="detail-label">文物描述</div>
            <div>{{ detail.linkedFind?.description || '-' }}</div>
          </div>
          <div>
            <div class="detail-label">送检时间</div>
            <div>{{ formatDateTime(detail.submittedAt) }}</div>
          </div>
          <div>
            <div class="detail-label">出结果时间</div>
            <div>{{ formatDateTime(detail.resultedAt) }}</div>
          </div>
        </div>

        <div class="result-box">
          <div class="detail-label">测年结果</div>
          <div v-if="detail.resultText" class="result-text">{{ detail.resultText }}</div>
          <div v-else class="muted">暂无结果</div>
        </div>

        <div v-if="detail.status === 'submitted'">
          <label class="result-input-label">
            登记测年结果 <span class="required">*</span>
            <textarea v-model="resultDraft" placeholder="填写实验室出具的测年结论，如校正年代区间、置信度等"></textarea>
          </label>
        </div>

        <p v-if="['resulted', 'void'].includes(detail.status)" class="lock-hint">
          当前为终态（{{ statusLabel(detail.status) }}），文物关联与单据内容不可再修改。
        </p>
        <p v-if="actionError" class="error">{{ actionError }}</p>

        <div class="modal-actions">
          <template v-if="detail.status === 'draft'">
            <button class="btn danger" @click="remove(detail)">删除</button>
            <button class="btn secondary" @click="openEdit(detail)">编辑</button>
            <button class="btn" @click="transition(detail, 'submit')">确认送检</button>
          </template>
          <template v-else-if="detail.status === 'submitted'">
            <button class="btn danger" @click="voidSubmission(detail)">作废</button>
            <button class="btn" @click="registerResult(detail)">登记结果</button>
          </template>
          <button class="btn secondary" @click="detail = null">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api/http'

const route = useRoute()

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

const showEditor = ref(false)
const formError = ref('')
const form = reactive({ id: null, labName: '', method: 'c14', linkedFindId: 0 })

const detail = ref(null)
const resultDraft = ref('')
const actionError = ref('')

function statusLabel(s) {
  return statusOptions.find((o) => o.value === s)?.label || s
}

function methodLabel(m) {
  return m === 'c14' ? '碳十四（C14）' : m === 'tl' ? '热释光（TL）' : m
}

function formatDate(v) {
  return v ? String(v).slice(0, 10) : '-'
}

function formatDateTime(v) {
  return v ? String(v).replace('T', ' ').slice(0, 16) : '-'
}

function findLocation(f) {
  if (!f) return '-'
  return `${f.unit?.site?.name || '-'} / ${f.unit?.code || '-'}`
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

function openCreate(presetFindId) {
  Object.assign(form, {
    id: null,
    labName: '',
    method: 'c14',
    linkedFindId: presetFindId || finds.value[0]?.id || 0
  })
  formError.value = ''
  detail.value = null
  showEditor.value = true
}

function openEdit(item) {
  Object.assign(form, {
    id: item.id,
    labName: item.labName,
    method: item.method,
    linkedFindId: item.linkedFindId || 0
  })
  formError.value = ''
  detail.value = null
  showEditor.value = true
}

async function save() {
  formError.value = ''
  if (!form.linkedFindId) {
    formError.value = '请选择关联的出土文物'
    return
  }
  if (!form.labName.trim()) {
    formError.value = '请填写承测实验室'
    return
  }
  const payload = {
    labName: form.labName.trim(),
    method: form.method,
    linkedFindId: form.linkedFindId
  }
  try {
    if (form.id) {
      await api.put(`/dating-submissions/${form.id}`, payload)
    } else {
      await api.post('/dating-submissions', payload)
    }
    showEditor.value = false
    await load()
  } catch (e) {
    formError.value = e.response?.data?.error || '保存失败'
  }
}

async function openDetail(item) {
  actionError.value = ''
  try {
    const { data } = await api.get(`/dating-submissions/${item.id}`)
    detail.value = data
    resultDraft.value = data.resultText || ''
  } catch (e) {
    error.value = e.response?.data?.error || '加载详情失败'
  }
}

async function transition(item, action, payload = {}) {
  actionError.value = ''
  try {
    const { data } = await api.post(`/dating-submissions/${item.id}/transition`, { action, ...payload })
    detail.value = data
    resultDraft.value = data.resultText || ''
    await load()
  } catch (e) {
    actionError.value = e.response?.data?.error || '流转失败'
  }
}

async function registerResult(item) {
  if (!resultDraft.value.trim()) {
    actionError.value = '请填写测年结果后再登记'
    return
  }
  await transition(item, 'result', { resultText: resultDraft.value.trim() })
}

async function voidSubmission(item) {
  if (!confirm('确认作废该送检单？作废后为终态，不可恢复。')) return
  await transition(item, 'void')
}

async function remove(item) {
  if (!confirm(`确认删除送检单 #${item.id}？仅草稿可删除。`)) return
  try {
    await api.delete(`/dating-submissions/${item.id}`)
    detail.value = null
    await load()
  } catch (e) {
    actionError.value = e.response?.data?.error || '删除失败'
  }
}

onMounted(async () => {
  await loadFinds()
  await load()
  if (route.query.findId) {
    const fid = Number(route.query.findId)
    if (finds.value.some((f) => f.id === fid)) {
      openCreate(fid)
    }
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
  min-width: 180px;
}

.muted {
  color: var(--muted);
  font-size: 0.88rem;
}

.find-link {
  color: var(--accent);
  font-weight: 600;
}

.required {
  color: var(--danger);
}

.status-tag {
  display: inline-block;
  padding: 0.15rem 0.6rem;
  border-radius: 999px;
  font-size: 0.8rem;
  background: #ece3d4;
  color: var(--muted);
}

.status-tag.st-draft {
  background: #ece3d4;
  color: #7a6a55;
}

.status-tag.st-submitted {
  background: #e3ecf5;
  color: #33567d;
}

.status-tag.st-resulted {
  background: #dcefe5;
  color: var(--ok);
}

.status-tag.st-void {
  background: #f3dcd7;
  color: var(--danger);
}

.editor-hint {
  margin: 0.75rem 0 0;
  font-size: 0.85rem;
  color: var(--muted);
}

.modal-wide {
  width: min(720px, 100%);
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem 1.25rem;
  margin-bottom: 1rem;
  font-size: 0.93rem;
}

.detail-label {
  font-size: 0.8rem;
  color: var(--muted);
  margin-bottom: 0.2rem;
}

.result-box {
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 0.85rem 1rem;
  background: #fbf6ee;
  margin-bottom: 1rem;
}

.result-text {
  white-space: pre-wrap;
  line-height: 1.6;
}

.result-input-label {
  margin-bottom: 0.5rem;
}

.lock-hint {
  color: var(--muted);
  font-size: 0.88rem;
  background: #f4ede2;
  border-radius: 8px;
  padding: 0.6rem 0.8rem;
}

@media (max-width: 800px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>
