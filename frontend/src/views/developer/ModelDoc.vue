<template>
  <div class="page-container">
    <el-page-header @back="$router.back()" :content="doc.name || '模型文档'" style="margin-bottom:20px" />

    <el-card class="quickstart" style="margin-bottom:20px">
      <template #header><strong>快速接入</strong></template>
      <el-descriptions :column="1" border size="small">
        <el-descriptions-item label="API 地址">POST {{ apiBase }}/v1/chat/completions</el-descriptions-item>
        <el-descriptions-item label="模型名称">{{ doc.model_id }}</el-descriptions-item>
        <el-descriptions-item label="认证方式">Authorization: Bearer sk-your-api-key</el-descriptions-item>
      </el-descriptions>
      <el-tabs style="margin-top:16px">
        <el-tab-pane label="curl">
          <pre class="code-block">curl {{ apiBase }}/v1/chat/completions \
  -H "Authorization: Bearer sk-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "{{ doc.model_id }}",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'</pre>
        </el-tab-pane>
        <el-tab-pane label="Python">
          <pre class="code-block">from openai import OpenAI

client = OpenAI(
    base_url="{{ apiBase }}/v1",
    api_key="sk-your-api-key"
)

response = client.chat.completions.create(
    model="{{ doc.model_id }}",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response.choices[0].message.content)</pre>
        </el-tab-pane>
        <el-tab-pane label="Node.js">
          <pre class="code-block">import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: '{{ apiBase }}/v1',
  apiKey: 'sk-your-api-key'
});

const response = await client.chat.completions.create({
  model: '{{ doc.model_id }}',
  messages: [{ role: 'user', content: 'Hello!' }]
});
console.log(response.choices[0].message.content);</pre>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-card v-if="doc.doc_content">
      <template #header><strong>接口文档</strong></template>
      <div class="doc-content" v-html="renderedDoc"></div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { developerApi } from '../../api/index.js'

const route = useRoute()
const doc = ref({ name: '', model_id: '', doc_content: '' })
const apiBase = computed(() => window.location.origin)
const renderedDoc = computed(() => {
  const text = doc.value.doc_content || ''
  return text.replace(/\n/g, '<br>').replace(/```([\s\S]*?)```/g, '<pre class="code-block"><code>$1</code></pre>')
})

onMounted(async () => {
  const res = await developerApi.getModelDoc(route.params.id)
  doc.value = res.data || {}
})
</script>

<style scoped>
.page-container { padding: 24px; }
.code-block { background: #1d1e2c; color: #e6e6e6; padding: 16px; border-radius: 6px; overflow-x: auto; font-family: 'Fira Code', Consolas, monospace; font-size: 13px; line-height: 1.5; white-space: pre; }
.doc-content { line-height: 1.8; font-size: 14px; }
</style>
