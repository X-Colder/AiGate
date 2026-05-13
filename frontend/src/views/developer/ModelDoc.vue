<template>
  <div class="page-container">
    <el-page-header @back="$router.back()" :content="modelName" style="margin-bottom:20px" />
    <el-card>
      <div class="doc-content" v-html="renderedDoc"></div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { developerApi } from '../../api/index.js'

const route = useRoute()
const doc = ref({ name: '', doc_content: '' })
const modelName = computed(() => doc.value.name || '模型文档')
const renderedDoc = computed(() => {
  const text = doc.value.doc_content || '暂无文档内容'
  return text.replace(/\n/g, '<br>').replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>')
})

onMounted(async () => {
  const res = await developerApi.getModelDoc(route.params.id)
  doc.value = res.data || {}
})
</script>

<style scoped>
.page-container { padding: 24px; }
.doc-content { line-height: 1.8; font-size: 14px; }
.doc-content pre { background: #f5f7fa; padding: 16px; border-radius: 4px; overflow-x: auto; }
.doc-content code { font-family: Consolas, monospace; font-size: 13px; }
</style>
