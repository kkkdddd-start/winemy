<template>
  <div class="module-view">
    <div class="module-header">
      <div class="header-info">
        <h2>编解码工具</h2>
        <p class="description">M25 - Base64/Hex/Unicode/URL/HTML</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="handleRefresh" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div class="feature-cards">
      <el-row :gutter="20">
        <el-col :span="4">
          <div class="feature-card" @click="handleEncode('base64')">
            <div class="card-icon">
              <el-icon><Key /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Base64</div>
              <div class="card-desc">Base64编解码</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="feature-card" @click="handleEncode('hex')">
            <div class="card-icon">
              <el-icon><List /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Hex</div>
              <div class="card-desc">十六进制编解码</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="feature-card" @click="handleEncode('unicode')">
            <div class="card-icon">
              <el-icon><ChatDotRound /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Unicode</div>
              <div class="card-desc">Unicode转换</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="feature-card" @click="handleEncode('url')">
            <div class="card-icon">
              <el-icon><Link /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">URL</div>
              <div class="card-desc">URL编码解码</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="feature-card" @click="handleEncode('html')">
            <div class="card-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">HTML</div>
              <div class="card-desc">HTML实体编解码</div>
            </div>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="feature-card" @click="handleEncode('hash')">
            <div class="card-icon">
              <el-icon><Lock /></el-icon>
            </div>
            <div class="card-content">
              <div class="card-title">Hash</div>
              <div class="card-desc">哈希计算</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="content-area">
      <el-card>
        <template #header>
          <div class="card-header">
            <span>编解码工具</span>
            <el-select v-model="currentCodec" placeholder="选择编解码类型" style="width: 200px">
              <el-option label="Base64" value="base64" />
              <el-option label="Hex" value="hex" />
              <el-option label="Unicode" value="unicode" />
              <el-option label="URL" value="url" />
              <el-option label="HTML" value="html" />
              <el-option label="Hash (MD5/SHA1/SHA256)" value="hash" />
            </el-select>
          </div>
        </template>
        <div class="codec-tool">
          <div class="codec-section">
            <div class="codec-label">输入</div>
            <el-input v-model="inputText" type="textarea" :rows="6" placeholder="请输入需要编码或解码的文本" />
          </div>
          <div class="codec-actions">
            <el-button type="primary" @click="handleEncode">编码/加密</el-button>
            <el-button @click="handleDecode">解码/解密</el-button>
            <el-button type="warning" @click="handleClear">清空</el-button>
          </div>
          <div class="codec-section">
            <div class="codec-label">输出</div>
            <el-input v-model="outputText" type="textarea" :rows="6" readonly placeholder="结果将显示在这里" />
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Refresh, Key, List, ChatDotRound, Link, Document, Lock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Go } from '@wailsjs/go/main/App'

const loading = ref(false)
const currentCodec = ref('base64')
const inputText = ref('')
const outputText = ref('')

function handleRefresh() {
  ElMessage.success('工具已刷新')
}

function handleEncode(type?: string) {
  const codec = type || currentCodec.value
  if (!inputText.value) {
    ElMessage.warning('请输入需要编码的文本')
    return
  }
  ElMessage.info(`${codec} 编码功能`)
}

function handleDecode() {
  if (!inputText.value) {
    ElMessage.warning('请输入需要解码的文本')
    return
  }
  ElMessage.info(`${currentCodec.value} 解码功能`)
}

function handleClear() {
  inputText.value = ''
  outputText.value = ''
}

onMounted(() => {
  // Initialize codec tool
})
</script>

<style scoped>
.module-view {
  height: 100%;
}

.module-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-info h2 {
  margin: 0 0 5px 0;
  font-size: 20px;
}

.description {
  margin: 0;
  color: #909399;
  font-size: 14px;
}

.feature-cards {
  margin-bottom: 20px;
}

.feature-card {
  background: #16213e;
  border-radius: 8px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 15px;
}

.feature-card:hover {
  background: #1a2a4a;
  transform: translateY(-2px);
}

.card-icon {
  width: 50px;
  height: 50px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  background: rgba(64, 158, 255, 0.2);
  color: #409eff;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 5px;
}

.card-desc {
  font-size: 12px;
  color: #909399;
}

.content-area {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.codec-tool {
  padding: 10px 0;
}

.codec-section {
  margin-bottom: 20px;
}

.codec-label {
  margin-bottom: 10px;
  font-size: 14px;
  color: #909399;
}

.codec-actions {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}
</style>
