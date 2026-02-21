<template>
  <div class="admin-settings p-4">
    <n-card title="系统设置" :bordered="false">
      <n-tabs type="line" animated>
        <!-- 基本设置 -->
        <n-tab-pane name="basic" tab="基本设置">
          <n-form
            ref="basicFormRef"
            :model="basic_form"
            label-placement="left"
            label-width="120px"
            class="mt-4"
            style="max-width: 600px;"
          >
            <n-form-item label="系统名称">
              <n-input v-model:value="basic_form.site_name" placeholder="请输入系统名称" />
            </n-form-item>
            <n-form-item label="系统描述">
              <n-input v-model:value="basic_form.site_desc" type="textarea" placeholder="请输入系统描述" :rows="3" />
            </n-form-item>
            <n-form-item label="版权信息">
              <n-input v-model:value="basic_form.copyright" placeholder="如: © 2024 F.st" />
            </n-form-item>
            <n-form-item label="ICP 备案号">
              <n-input v-model:value="basic_form.icp" placeholder="如: 京ICP备xxxxx号" />
            </n-form-item>
            <n-form-item label="用户注册">
              <n-switch v-model:value="basic_form.allow_register" />
              <n-text class="ml-2" depth="3">{{ basic_form.allow_register ? '允许新用户注册' : '禁止新用户注册' }}</n-text>
            </n-form-item>
            <n-form-item>
              <n-button type="primary" @click="handle_save_basic" :loading="saving">
                保存设置
              </n-button>
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <!-- 邮件设置 -->
        <n-tab-pane name="email" tab="邮件设置">
          <n-form
            :model="email_form"
            label-placement="left"
            label-width="120px"
            class="mt-4"
            style="max-width: 600px;"
          >
            <n-form-item label="SMTP 服务器">
              <n-input v-model:value="email_form.smtp_host" placeholder="如: smtp.gmail.com" />
            </n-form-item>
            <n-form-item label="SMTP 端口">
              <n-input-number v-model:value="email_form.smtp_port" :min="1" :max="65535" style="width: 100%;" />
            </n-form-item>
            <n-form-item label="发件人邮箱">
              <n-input v-model:value="email_form.smtp_username" placeholder="发件人邮箱地址" />
            </n-form-item>
            <n-form-item label="邮箱密码">
              <n-input v-model:value="email_form.smtp_password" type="password" show-password-on="click" placeholder="邮箱密码或应用密钥" />
            </n-form-item>
            <n-form-item label="SSL 加密">
              <n-switch v-model:value="email_form.smtp_ssl" />
            </n-form-item>
            <n-form-item>
              <n-space>
                <n-button type="primary" @click="handle_save_email" :loading="saving">保存</n-button>
                <n-button @click="handle_test_email">发送测试邮件</n-button>
              </n-space>
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <!-- 安全设置 -->
        <n-tab-pane name="security" tab="安全设置">
          <n-form
            :model="security_form"
            label-placement="left"
            label-width="160px"
            class="mt-4"
            style="max-width: 600px;"
          >
            <n-form-item label="极验验证码">
              <n-switch v-model:value="security_form.geetest_enabled" />
              <n-text class="ml-2" depth="3">{{ security_form.geetest_enabled ? '已启用' : '已禁用' }}</n-text>
            </n-form-item>
            <n-form-item v-if="security_form.geetest_enabled" label="极验 ID">
              <n-input v-model:value="security_form.geetest_id" placeholder="Geetest Captcha ID" />
            </n-form-item>
            <n-form-item v-if="security_form.geetest_enabled" label="极验 Key">
              <n-input v-model:value="security_form.geetest_key" type="password" show-password-on="click" placeholder="Geetest Key" />
            </n-form-item>
            <n-divider />
            <n-form-item label="JWT Token 有效期 (秒)">
              <n-input-number v-model:value="security_form.jwt_expire" :min="300" :step="300" style="width: 100%;" />
            </n-form-item>
            <n-form-item label="Refresh Token 有效期 (秒)">
              <n-input-number v-model:value="security_form.refresh_expire" :min="3600" :step="3600" style="width: 100%;" />
            </n-form-item>
            <n-form-item label="登录失败锁定次数">
              <n-input-number v-model:value="security_form.max_login_attempts" :min="3" :max="20" style="width: 100%;" />
            </n-form-item>
            <n-form-item>
              <n-button type="primary" @click="handle_save_security" :loading="saving">
                保存设置
              </n-button>
            </n-form-item>
          </n-form>
        </n-tab-pane>

        <!-- 系统信息 -->
        <n-tab-pane name="info" tab="系统信息">
          <div class="mt-4">
            <n-descriptions bordered :column="2" label-placement="left">
              <n-descriptions-item label="系统版本">v1.0.0</n-descriptions-item>
              <n-descriptions-item label="Go 版本">1.24+</n-descriptions-item>
              <n-descriptions-item label="前端框架">Vue 3 + Vite</n-descriptions-item>
              <n-descriptions-item label="UI 组件库">Naive UI</n-descriptions-item>
              <n-descriptions-item label="数据库">MySQL 8.0+</n-descriptions-item>
              <n-descriptions-item label="运行环境">{{ mode }}</n-descriptions-item>
              <n-descriptions-item label="Node.js">18+</n-descriptions-item>
              <n-descriptions-item label="构建时间">{{ build_time }}</n-descriptions-item>
            </n-descriptions>

            <n-card title="已加载插件" class="mt-4" size="small">
              <n-space vertical>
                <n-tag v-for="p in plugins" :key="p.name" :type="p.active ? 'success' : 'default'" size="medium">
                  {{ p.name }} ({{ p.version }})
                </n-tag>
              </n-space>
            </n-card>
          </div>
        </n-tab-pane>
      </n-tabs>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import {
  NCard, NTabs, NTabPane, NForm, NFormItem, NInput, NInputNumber,
  NButton, NSwitch, NSpace, NText, NDivider, NDescriptions,
  NDescriptionsItem, NTag, useMessage,
} from 'naive-ui'

const message = useMessage()
const saving = ref(false)
const mode = import.meta.env.MODE
const build_time = typeof __BUILD_TIMESTAMP__ !== 'undefined' ? __BUILD_TIMESTAMP__ : '开发模式'

// 基本设置
const basic_form = reactive({
  site_name: 'F.st',
  site_desc: '基于 Go + Vue 3 的全栈管理系统模板',
  copyright: '© 2024 F.st',
  icp: '',
  allow_register: true,
})

// 邮件设置
const email_form = reactive({
  smtp_host: '',
  smtp_port: 587,
  smtp_username: '',
  smtp_password: '',
  smtp_ssl: true,
})

// 安全设置
const security_form = reactive({
  geetest_enabled: false,
  geetest_id: '',
  geetest_key: '',
  jwt_expire: 7200,
  refresh_expire: 604800,
  max_login_attempts: 5,
})

// 插件列表 (模拟)
const plugins = ref([
  { name: 'Demo Plugin', version: '1.0.0', active: true },
  { name: 'Email', version: '1.0.0', active: true },
])

async function handle_save_basic() {
  saving.value = true
  try {
    // TODO: 对接 API
    await new Promise(r => setTimeout(r, 500))
    message.success('基本设置保存成功')
  } catch {
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

async function handle_save_email() {
  saving.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    message.success('邮件设置保存成功')
  } catch {
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

function handle_test_email() {
  message.info('发送测试邮件…（开发中）')
}

async function handle_save_security() {
  saving.value = true
  try {
    await new Promise(r => setTimeout(r, 500))
    message.success('安全设置保存成功')
  } catch {
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.admin-settings { padding: 16px; }
.mt-4 { margin-top: 16px; }
.ml-2 { margin-left: 8px; }
</style>
