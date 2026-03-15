<template>
  <n-card title="系统设置" :bordered="false">
    <n-spin :show="loading">
      <n-tabs v-model:value="topTab" type="line" animated>
        <n-tab-pane name="system-config" tab="系统配置">
          <n-tabs v-model:value="systemSubTab" type="line" placement="left" animated>
            <n-tab-pane name="basic" tab="基本设置">
              <n-space vertical>
                <n-form :model="basicForm" label-placement="left" label-width="120px" style="max-width: 640px;">
                  <n-form-item label="系统名称">
                    <n-input v-model:value="basicForm.site_name" placeholder="请输入系统名称" />
                  </n-form-item>
                  <n-form-item label="系统描述">
                    <n-input v-model:value="basicForm.site_desc" type="textarea" placeholder="请输入系统描述" :rows="3" />
                  </n-form-item>
                  <n-form-item label="站点Logo">
                    <n-input v-model:value="basicForm.site_logo" placeholder="Logo图片URL" />
                  </n-form-item>
                  <n-form-item label="版权信息">
                    <n-input v-model:value="basicForm.copyright" placeholder="如: (c) 2024 F.st" />
                  </n-form-item>
                  <n-form-item label="ICP 备案号">
                    <n-input v-model:value="basicForm.icp" placeholder="如: 京ICP备xxxxx号" />
                  </n-form-item>
                  <n-form-item label="系统版本">
                    <n-input v-model:value="basicForm.version" placeholder="如: 1.0.0" />
                  </n-form-item>
                  <n-form-item label="默认语言">
                    <n-select v-model:value="basicForm.default_lang" :options="langOptions" placeholder="选择默认语言" />
                  </n-form-item>
                  <n-form-item label="前端地址">
                    <n-input v-model:value="basicForm.frontend_url" placeholder="如: http://example.com（结尾不要加 /）" />
                  </n-form-item>
                  <n-form-item label="后端API地址">
                    <n-input v-model:value="basicForm.backend_api_url" placeholder="如: http://api.example.com（结尾不要加 /）" />
                  </n-form-item>
                  <n-form-item label="用户注册">
                    <n-space align="center">
                      <n-switch
                        :value="basicForm.allow_register"
                        :loading="switchLoading.allow_register"
                        @update:value="handleUpdateAllowRegister"
                      />
                      <n-text depth="3">{{ basicForm.allow_register ? '允许新用户注册' : '禁止新用户注册' }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item>
                    <n-button type="primary" :loading="savingBasic" @click="handleSaveBasic">保存设置</n-button>
                  </n-form-item>
                </n-form>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="email" tab="邮件设置">
              <n-space vertical>
                <n-form :model="emailForm" label-placement="left" label-width="120px" style="max-width: 640px;">
                  <n-form-item label="邮箱验证码">
                    <n-space align="center">
                      <n-switch
                        :value="emailForm.email_verify_enabled"
                        :loading="switchLoading.email_verify_enabled"
                        @update:value="handleUpdateEmailVerifyEnabled"
                      />
                      <n-text depth="3">{{ emailForm.email_verify_enabled ? '已启用（修改邮箱需验证码）' : '已禁用（修改邮箱直接生效）' }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-divider />
                  <n-form-item label="SMTP 服务器">
                    <n-input v-model:value="emailForm.smtp_host" placeholder="如: smtp.gmail.com" />
                  </n-form-item>
                  <n-form-item label="SMTP 端口">
                    <n-input-number v-model:value="emailForm.smtp_port" :min="1" :max="65535" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item label="发件人邮箱">
                    <n-input v-model:value="emailForm.smtp_username" placeholder="发件人邮箱地址" />
                  </n-form-item>
                  <n-form-item label="邮箱密码">
                    <n-input
                      v-model:value="emailForm.smtp_password"
                      type="password"
                      show-password-on="click"
                      placeholder="邮箱密码或应用密钥"
                    />
                  </n-form-item>
                  <n-form-item label="发件人名称">
                    <n-input v-model:value="emailForm.system_email_name" placeholder="如: F.st" />
                  </n-form-item>
                  <n-form-item label="SSL 加密">
                    <n-space align="center">
                      <n-switch :value="emailForm.smtp_ssl" :loading="switchLoading.smtp_ssl" @update:value="handleUpdateSmtpSSL" />
                      <n-text depth="3">{{ emailForm.smtp_ssl ? '已启用 SSL' : '未启用 SSL' }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item>
                    <n-space>
                      <n-button type="primary" :loading="savingEmail" @click="handleSaveEmail">保存</n-button>
                      <n-button :loading="testingEmail" @click="handleTestEmail">发送测试邮件</n-button>
                    </n-space>
                  </n-form-item>
                </n-form>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="sms" tab="短信设置">
              <n-space vertical>
                <n-form :model="smsForm" label-placement="left" label-width="120px" style="max-width: 640px;">
                  <n-form-item label="短信验证码">
                    <n-space align="center">
                      <n-switch
                        :value="smsForm.sms_verify_enabled"
                        :loading="switchLoading.sms_verify_enabled"
                        @update:value="handleUpdateSmsVerifyEnabled"
                      />
                      <n-text depth="3">{{ smsForm.sms_verify_enabled ? '已启用（修改手机号需验证码）' : '已禁用（修改手机号直接生效）' }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-divider />
                  <n-form-item label="短信服务商">
                    <n-select
                      v-model:value="smsForm.sms_provider"
                      :options="smsProviderOptions"
                      placeholder="选择短信服务商"
                    />
                  </n-form-item>
                  <n-form-item label="AccessKey">
                    <n-input v-model:value="smsForm.sms_access_key" placeholder="短信服务商 AccessKey" />
                  </n-form-item>
                  <n-form-item label="SecretKey">
                    <n-input
                      v-model:value="smsForm.sms_secret_key"
                      type="password"
                      show-password-on="click"
                      placeholder="短信服务商 SecretKey"
                    />
                  </n-form-item>
                  <n-form-item label="短信签名">
                    <n-input v-model:value="smsForm.sms_sign_name" placeholder="如: F.st" />
                  </n-form-item>
                  <n-form-item label="验证码模板ID">
                    <n-input v-model:value="smsForm.sms_template_code" placeholder="短信验证码模板ID" />
                  </n-form-item>
                  <n-form-item label="服务区域">
                    <n-input v-model:value="smsForm.sms_region" placeholder="部分服务商需要，如: cn-hangzhou" />
                  </n-form-item>
                  <n-form-item>
                    <n-button type="primary" :loading="savingSms" @click="handleSaveSms">保存设置</n-button>
                  </n-form-item>
                </n-form>
                <n-alert type="info" title="提示" :bordered="false">
                  当前仅 <n-text strong>console</n-text> 模式可用（验证码打印到后端控制台日志）。阿里云、腾讯云等服务商接入后即可切换。
                </n-alert>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="security" tab="安全设置">
              <n-space vertical>
                <n-form :model="securityForm" label-placement="left" label-width="180px" style="max-width: 640px;">
                  <n-form-item label="极验验证码">
                    <n-space align="center">
                      <n-switch
                        :value="securityForm.geetest_enabled"
                        :loading="switchLoading.geetest_enabled"
                        @update:value="handleUpdateGeetestEnabled"
                      />
                      <n-text depth="3">{{ securityForm.geetest_enabled ? '已启用' : '已禁用' }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item label="账号注销">
                    <n-space align="center">
                      <n-switch
                        :value="securityForm.allow_delete_account"
                        :loading="switchLoading.allow_delete_account"
                        @update:value="handleUpdateAllowDeleteAccount"
                      />
                      <n-text depth="3">{{ securityForm.allow_delete_account ? '允许用户主动注销账号' : '禁止用户主动注销账号' }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-form-item label="极验 Captcha ID">
                    <n-input v-model:value="securityForm.geetest_captcha_id" placeholder="请输入极验验证码 ID" />
                  </n-form-item>
                  <n-form-item label="极验 Captcha Key">
                    <n-input
                      v-model:value="securityForm.geetest_captcha_key"
                      type="password"
                      show-password-on="click"
                      placeholder="请输入极验验证码 Key"
                    />
                  </n-form-item>
                  <n-divider />
                  <n-form-item label="JWT Token 有效期 (秒)">
                    <n-input-number v-model:value="securityForm.jwt_access_expire" :min="300" :step="300" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item label="Refresh Token 有效期 (秒)">
                    <n-input-number v-model:value="securityForm.jwt_refresh_expire" :min="3600" :step="3600" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item label="登录失败锁定次数">
                    <n-input-number v-model:value="securityForm.login_max_failure" :min="3" :max="20" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item label="账户锁定时长 (分钟)">
                    <n-input-number v-model:value="securityForm.login_lock_duration" :min="1" :max="1440" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item>
                    <n-space>
                      <n-button type="primary" :loading="savingSecurity" @click="handleSaveSecurity">保存设置</n-button>
                      <n-button type="warning" :loading="restartingBackend" @click="handleRestartBackend">重启后端</n-button>
                    </n-space>
                  </n-form-item>
                </n-form>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="payment" tab="支付设置">
              <n-space vertical>
                <n-form :model="paymentForm" label-placement="left" label-width="140px" style="max-width: 640px;">
                  <n-form-item label="支付功能">
                    <n-space align="center">
                      <n-switch
                        :value="paymentForm.payment_enabled"
                        :loading="switchLoading.payment_enabled"
                        @update:value="handleUpdatePaymentEnabled"
                      />
                      <n-text depth="3">{{ paymentForm.payment_enabled ? '已启用' : '已禁用' }}</n-text>
                    </n-space>
                  </n-form-item>
                  <n-divider />
                  <n-form-item label="订单有效期（分钟）">
                    <n-input-number v-model:value="paymentForm.payment_order_expire_minutes" :min="1" :max="1440" style="width: 100%;" />
                  </n-form-item>
                  <n-form-item>
                    <n-button type="primary" :loading="savingPayment" @click="handleSavePayment">保存设置</n-button>
                  </n-form-item>
                </n-form>
                <n-alert type="info" title="配置说明" :bordered="false">
                  <ul style="margin: 0; padding-left: 18px;">
                    <li>启用「支付功能」后，用户端会展示所有已启用的支付通道</li>
                    <li>每个通道的网关地址、商户ID、密钥等信息在「支付渠道」页面配置</li>
                  </ul>
                </n-alert>
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="custom" tab="自定义配置">
              <n-space vertical :size="16">
                <n-space justify="end">
                  <n-button type="primary" @click="showAddModal = true">
                    <template #icon>
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" style="width: 1em; height: 1em;">
                        <path d="M11 11V5h2v6h6v2h-6v6h-2v-6H5v-2z" />
                      </svg>
                    </template>
                    添加配置项
                  </n-button>
                </n-space>

                <n-data-table :columns="customColumns" :data="customSettings" :pagination="false" :bordered="false" />
              </n-space>
            </n-tab-pane>
          </n-tabs>
        </n-tab-pane>

        <n-tab-pane name="email-templates" tab="邮件模板">
          <EmailTemplates />
        </n-tab-pane>

        <n-tab-pane name="operation-logs" tab="操作日志">
          <OperationLogs />
        </n-tab-pane>

        <n-tab-pane name="info" tab="系统信息">
          <n-space vertical>
            <n-descriptions bordered :column="2" label-placement="left">
              <n-descriptions-item label="系统版本">{{ settingsStore.version }}</n-descriptions-item>
              <n-descriptions-item label="Go 版本">1.24+</n-descriptions-item>
              <n-descriptions-item label="前端框架">Vue 3 + Vite</n-descriptions-item>
              <n-descriptions-item label="UI 组件库">Naive UI</n-descriptions-item>
              <n-descriptions-item label="数据库">MySQL 8.0+</n-descriptions-item>
              <n-descriptions-item label="运行环境">{{ mode }}</n-descriptions-item>
              <n-descriptions-item label="Node.js">18+</n-descriptions-item>
              <n-descriptions-item label="构建时间">{{ buildTime }}</n-descriptions-item>
            </n-descriptions>

            <n-card title="已加载插件" size="small">
              <n-space vertical>
                <n-tag v-for="p in plugins" :key="p.name" :type="p.active ? 'success' : 'default'" size="medium">
                  {{ p.name }} ({{ p.version }})
                </n-tag>
              </n-space>
            </n-card>
          </n-space>
        </n-tab-pane>

        <n-tab-pane name="server-management" tab="服务器管理">
          <n-tabs type="line" animated>
            <n-tab-pane name="monitor" tab="📊 系统监控">
              <n-space vertical :size="16">
                <n-space justify="space-between" align="center">
                  <n-text depth="3">实时监控数据</n-text>
                  <n-button :loading="loadingServerMonitoring" @click="loadServerMonitoringStatus">刷新</n-button>
                </n-space>

                <n-grid :x-gap="10" :y-gap="10" cols="2 s:4 m:4 l:8" responsive="screen">
                  <n-gi>
                    <n-card size="small">
                      <n-statistic label="CPU">
                        <template #default>
                          <n-text :type="cpuPercent > 80 ? 'error' : 'success'">{{ formatPercent(cpuPercent) }}</n-text>
                        </template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">{{ serverMonitoringData?.metrics.cpu.core_count || 0 }}核</n-text></template>
                      </n-statistic>
                      <n-progress type="line" :percentage="cpuPercent" :status="cpuPercent > 80 ? 'error' : 'success'" :show-indicator="false" style="margin-top: 8px" />
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic label="系统内存">
                        <template #default>
                          <n-text :type="memoryPercent > 80 ? 'error' : 'success'">{{ formatPercent(memoryPercent) }}</n-text>
                        </template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">{{ formatStorageFromMB(serverMonitoringData?.metrics.memory.used_mb || 0) }}/{{ formatStorageFromMB(serverMonitoringData?.metrics.memory.total_mb || 0) }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic label="Swap">
                        <template #default>
                          <n-text :type="swapPercent > 80 ? 'error' : 'success'">{{ formatPercent(swapPercent) }}</n-text>
                        </template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">{{ formatStorageFromMB(serverMonitoringData?.metrics.swap.used_mb || 0) }}/{{ formatStorageFromMB(serverMonitoringData?.metrics.swap.total_mb || 0) }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic label="进程内存">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.process_rss_mb || 0) }}</template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">CPU {{ Number((serverMonitoringData?.process.process_cpu || 0).toFixed(2)) }}%</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic label="Go 堆内存">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.heap_alloc_mb || 0) }}</template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">sys {{ formatStorageFromMB(serverMonitoringData?.process.memory_sys_mb || 0) }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic label="协程/GC">
                        <template #default>{{ serverMonitoringData?.process.goroutines || 0 }}</template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">GC {{ serverMonitoringData?.process.gc_count || 0 }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic label="磁盘使用率">
                        <template #default>
                          <n-text :type="diskPercent > 80 ? 'error' : 'success'">{{ formatPercent(diskPercent) }}</n-text>
                        </template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">{{ formatStorageFromGB(serverMonitoringData?.metrics.disk.used_gb || 0) }}/{{ formatStorageFromGB(serverMonitoringData?.metrics.disk.total_gb || 0) }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-statistic label="运行时间">
                        <template #default>{{ uptimeTextPrecise }}</template>
                        <template #suffix><n-text depth="3" style="font-size: 10px">启动: {{ startTimeText }} · {{ uptimeText }}</n-text></template>
                      </n-statistic>
                    </n-card>
                  </n-gi>
                  <n-gi>
                    <n-card size="small">
                      <n-space vertical size="small">
                        <n-statistic label="网络">
                          <template #default>{{ formatBytes((serverMonitoringData?.metrics.network.bytes_sent || 0) + (serverMonitoringData?.metrics.network.bytes_recv || 0)) }}</template>
                        </n-statistic>
                        <n-space justify="space-between"><n-text depth="3">上传</n-text><n-text>{{ formatBytes(serverMonitoringData?.metrics.network.bytes_sent || 0) }}</n-text></n-space>
                        <n-space justify="space-between"><n-text depth="3">下载</n-text><n-text>{{ formatBytes(serverMonitoringData?.metrics.network.bytes_recv || 0) }}</n-text></n-space>
                        <n-space justify="space-between"><n-text depth="3">上传包</n-text><n-text>{{ formatInteger(serverMonitoringData?.metrics.network.packets_sent || 0) }}</n-text></n-space>
                        <n-space justify="space-between"><n-text depth="3">下载包</n-text><n-text>{{ formatInteger(serverMonitoringData?.metrics.network.packets_recv || 0) }}</n-text></n-space>
                      </n-space>
                    </n-card>
                  </n-gi>
                </n-grid>

                <n-card size="small" title="💾 内存详情">
                  <n-grid :x-gap="10" :y-gap="10" cols="1 s:2 m:4 l:4" responsive="screen">
                    <n-gi>
                      <n-statistic label="Go内存分配">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.memory_alloc_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic label="Go内存系统">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.memory_sys_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic label="堆分配">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.heap_alloc_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic label="堆使用中">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.heap_inuse_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic label="堆空闲">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.heap_idle_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic label="栈使用">
                        <template #default>{{ formatStorageFromMB(serverMonitoringData?.process.stack_inuse_mb || 0) }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic label="GC次数">
                        <template #default>{{ serverMonitoringData?.process.gc_count || 0 }}</template>
                      </n-statistic>
                    </n-gi>
                    <n-gi>
                      <n-statistic label="GC CPU占用">
                        <template #default>{{ Number(((serverMonitoringData?.process.gc_cpu_fraction || 0) * 100).toFixed(2)) }}%</template>
                      </n-statistic>
                    </n-gi>
                  </n-grid>
                </n-card>

                <n-space justify="space-between" align="center">
                  <n-text depth="3">服务健康快照</n-text>
                  <n-text depth="3" style="font-size: 12px">最近刷新: {{ serverMonitoringGeneratedAt || '-' }}</n-text>
                </n-space>

                <n-data-table
                  :columns="serviceStatusColumns"
                  :data="serviceStatusRows"
                  :pagination="false"
                  :bordered="false"
                />
              </n-space>
            </n-tab-pane>

            <n-tab-pane name="debug" tab="🔬 调试工具">
              <n-space vertical :size="16">
                <n-card title="📊 系统概览" size="small">
                  <template #header-extra>
                    <n-space>
                      <n-button size="small" :type="debugAutoRefresh ? 'primary' : 'default'" @click="toggleDebugAutoRefresh(!debugAutoRefresh)">
                        {{ debugAutoRefresh ? '停止刷新' : '自动刷新' }}
                      </n-button>
                      <n-button size="small" :loading="loadingDebugStats" @click="loadDebugStats">刷新</n-button>
                      <n-button size="small" type="warning" @click="handleForceGC">强制GC</n-button>
                    </n-space>
                  </template>

                  <n-grid :x-gap="12" :y-gap="12" cols="1 s:2 m:2 l:2" responsive="screen">
                    <n-gi>
                      <n-card size="small" title="进程资源">
                        <n-space vertical size="small">
                          <div>
                            <n-space justify="space-between"><n-text>CPU</n-text><n-text>{{ Number((serverMonitoringData?.process.process_cpu || 0).toFixed(1)) }}%</n-text></n-space>
                            <n-progress type="line" :percentage="Number((serverMonitoringData?.process.process_cpu || 0).toFixed(1))" :status="(serverMonitoringData?.process.process_cpu ?? 0) > 80 ? 'error' : 'success'" :show-indicator="false" style="margin-top: 4px" />
                          </div>
                          <n-space justify="space-between"><n-text>内存</n-text><n-text>{{ formatStorageFromMB(serverMonitoringData?.process.process_rss_mb || 0) }}</n-text></n-space>
                          <n-space justify="space-between"><n-text>协程</n-text><n-text>{{ serverMonitoringData?.process.goroutines || 0 }}</n-text></n-space>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" title="协程统计">
                        <n-space vertical size="small">
                          <n-space justify="space-between"><n-text>运行时总数</n-text><n-text>{{ debugStats?.total_count || 0 }}</n-text></n-space>
                          <n-space justify="space-between"><n-text>已跟踪</n-text><n-text>{{ debugStats?.tracked_count || 0 }}</n-text></n-space>
                          <n-space justify="space-between"><n-text>潜在泄漏</n-text><n-text type="error">{{ debugStats?.potential_leaks || 0 }}</n-text></n-space>
                        </n-space>
                      </n-card>
                    </n-gi>
                  </n-grid>
                </n-card>

                <n-card title="🔬 性能分析 (pprof)" size="small">
                  <template #header-extra>
                    <n-button size="small" @click="clearAllPprofResults">清空结果</n-button>
                  </template>

                  <n-grid :x-gap="12" :y-gap="12" cols="1 s:2 m:3 l:3" responsive="screen">
                    <n-gi>
                      <n-card size="small" title="CPU Profile">
                        <n-space vertical size="small">
                          <n-text depth="3">采集 CPU 热点</n-text>
                          <n-space>
                            <n-input-number v-model:value="pprofConfig.cpuSeconds" :min="5" :max="120" size="small" style="width: 90px" />
                            <n-button size="small" type="primary" :loading="pprofLoading.cpu" @click="captureCPUProfile">采集</n-button>
                          </n-space>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" title="Heap (内存)">
                        <n-space vertical size="small">
                          <n-text depth="3">采集堆内存分配</n-text>
                          <n-button size="small" type="primary" :loading="pprofLoading.heap" @click="captureHeapProfile">采集</n-button>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" title="Goroutine">
                        <n-space vertical size="small">
                          <n-text depth="3">采集协程堆栈</n-text>
                          <n-button size="small" type="primary" :loading="pprofLoading.goroutine" @click="captureGoroutineProfile">采集</n-button>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" title="Allocs (分配)">
                        <n-space vertical size="small">
                          <n-text depth="3">采集内存分配采样</n-text>
                          <n-button size="small" type="primary" :loading="pprofLoading.allocs" @click="captureAllocsProfile">采集</n-button>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" title="Block (阻塞)">
                        <n-space vertical size="small">
                          <n-text depth="3">采集阻塞事件</n-text>
                          <n-button size="small" type="primary" :loading="pprofLoading.block" @click="captureBlockProfile">采集</n-button>
                        </n-space>
                      </n-card>
                    </n-gi>
                    <n-gi>
                      <n-card size="small" title="Mutex (互斥锁)">
                        <n-space vertical size="small">
                          <n-text depth="3">采集互斥锁竞争</n-text>
                          <n-button size="small" type="primary" :loading="pprofLoading.mutex" @click="captureMutexProfile">采集</n-button>
                        </n-space>
                      </n-card>
                    </n-gi>
                  </n-grid>

                  <n-empty v-if="!hasAnyPprofResult" description="点击上方按钮采集性能数据" style="margin-top: 16px" />
                  <n-space v-else vertical :size="12" style="margin-top: 16px">
                    <n-card v-if="pprofResults.cpu" size="small" title="CPU Profile 结果">
                      <n-code :code="pprofResults.cpuText || ''" language="text" word-wrap />
                    </n-card>
                    <n-card v-if="pprofResults.heap" size="small" title="Heap Profile 结果">
                      <n-code :code="pprofResults.heapText || ''" language="text" word-wrap />
                    </n-card>
                    <n-card v-if="pprofResults.goroutine" size="small" title="Goroutine Profile 结果">
                      <n-code :code="pprofResults.goroutine || ''" language="text" word-wrap />
                    </n-card>
                    <n-card v-if="pprofResults.allocs" size="small" title="Allocs Profile 结果">
                      <n-code :code="pprofResults.allocsText || ''" language="text" word-wrap />
                    </n-card>
                    <n-card v-if="pprofResults.block" size="small" title="Block Profile 结果">
                      <n-code :code="pprofResults.blockText || ''" language="text" word-wrap />
                    </n-card>
                    <n-card v-if="pprofResults.mutex" size="small" title="Mutex Profile 结果">
                      <n-code :code="pprofResults.mutexText || ''" language="text" word-wrap />
                    </n-card>
                  </n-space>
                </n-card>

                <n-card title="📚 运行时协程堆栈" size="small">
                  <template #header-extra>
                    <n-space>
                      <n-tooltip trigger="hover">
                        <template #trigger>
                          <n-input-number v-model:value="stackFilterMinWaitMinutes" :min="0" size="small" style="width: 140px" placeholder="最小等待分钟" />
                        </template>
                        过滤显示等待时间超过指定分钟数的协程，0表示显示全部
                      </n-tooltip>
                      <n-button size="small" :loading="loadingRuntimeStacks" @click="loadRuntimeStacks">加载堆栈</n-button>
                      <n-button size="small" @click="clearRuntimeStacks">清空堆栈</n-button>
                    </n-space>
                  </template>
                  <n-empty v-if="runtimeStackText === ''" description="点击加载堆栈查看运行时协程信息" />
                  <n-code v-else :code="runtimeStackText" language="text" word-wrap />
                </n-card>
              </n-space>
            </n-tab-pane>
          </n-tabs>
        </n-tab-pane>
      </n-tabs>
    </n-spin>

    <n-modal v-model:show="showAddModal" preset="card" title="添加配置项" style="width: 500px;" :mask-closable="false">
      <n-form ref="addFormRef" :model="addForm" :rules="addFormRules" label-placement="left" label-width="100px">
        <n-form-item label="配置键名" path="key">
          <n-input v-model:value="addForm.key" placeholder="如: custom_field" />
        </n-form-item>
        <n-form-item label="配置值" path="value">
          <n-input v-model:value="addForm.value" placeholder="配置值" />
        </n-form-item>
        <n-form-item label="显示名称" path="label">
          <n-input v-model:value="addForm.label" placeholder="如: 自定义字段" />
        </n-form-item>
        <n-form-item label="值类型" path="type">
          <n-select v-model:value="addForm.type" :options="typeOptions" />
        </n-form-item>
        <n-form-item label="描述说明" path="description">
          <n-input v-model:value="addForm.description" placeholder="配置项说明" />
        </n-form-item>
        <n-form-item label="是否公开">
          <n-switch v-model:value="addForm.is_public" />
          <n-text depth="3" style="margin-left: 8px;">公开配置可被前端获取</n-text>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showAddModal = false">取消</n-button>
          <n-button type="primary" :loading="adding" @click="handleAddSetting">添加</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showEditModal" preset="card" title="修改配置项" style="width: 520px;" :mask-closable="false">
      <n-form label-placement="left" label-width="100px">
        <n-form-item label="配置键名">
          <n-input v-model:value="editForm.key" disabled />
        </n-form-item>
        <n-form-item label="配置值">
          <n-input v-model:value="editForm.value" placeholder="配置值" />
        </n-form-item>
        <n-form-item label="显示名称">
          <n-input v-model:value="editForm.label" placeholder="显示名称" />
        </n-form-item>
        <n-form-item label="值类型">
          <n-select v-model:value="editForm.type" :options="typeOptions" />
        </n-form-item>
        <n-form-item label="描述说明">
          <n-input v-model:value="editForm.description" placeholder="配置项说明" />
        </n-form-item>
        <n-form-item label="是否公开">
          <n-switch v-model:value="editForm.is_public" />
          <n-text depth="3" style="margin-left: 8px;">公开配置可被前端获取</n-text>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showEditModal = false">取消</n-button>
          <n-button type="primary" :loading="savingEdit" @click="handleSaveSettingEdit">保存</n-button>
        </n-space>
      </template>
    </n-modal>

  </n-card>
</template>

<script setup lang="ts">
import { computed, h, onMounted, onUnmounted, reactive, ref } from 'vue'
import {
  NAlert,
  NButton,
  NCode,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NDivider,
  NForm,
  NFormItem,
  NGi,
  NGrid,
  NInput,
  NInputNumber,
  NModal,
  NProgress,
  NSelect,
  NSpace,
  NSpin,
  NStatistic,
  NSwitch,
  NTabPane,
  NTabs,
  NTag,
  NText,
  NTooltip,
  type DataTableColumns,
  useMessage,
} from 'naive-ui'
import { adminApi } from '@/service/api/admin'
import { adminDebugApi } from '@/service/api/admin/debug'
import EmailTemplates from '@/views/admin/email-templates/index.vue'
import OperationLogs from '@/views/admin/logs/index.vue'
import type { ServerMonitoringStatusResponse, SettingDTO, SettingType } from '@/service/api/admin/settings'
import { useSettingsStore } from '@/store/settings'
import { local } from '@/utils'

const message = useMessage()
const settingsStore = useSettingsStore()

const loading = ref(true)
const adding = ref(false)
const savingEdit = ref(false)
const showAddModal = ref(false)
const showEditModal = ref(false)
const savingBasic = ref(false)
const savingEmail = ref(false)
const savingSecurity = ref(false)
const savingPayment = ref(false)
const testingEmail = ref(false)
const restartingBackend = ref(false)
const loadingServerMonitoring = ref(false)
const topTab = ref('system-config')
const systemSubTab = ref('basic')
const mode = import.meta.env.MODE
const buildTime = typeof __BUILD_TIMESTAMP__ !== 'undefined' ? __BUILD_TIMESTAMP__ : '开发模式'
const serverMonitoringGeneratedAt = ref('')
const serverMonitoringData = ref<ServerMonitoringStatusResponse | null>(null)

const loadingDebugStats = ref(false)
const debugAutoRefresh = ref(false)
const debugRefreshInterval = ref<number | null>(null)
const debugStats = ref<any>(null)
const loadingRuntimeStacks = ref(false)
const runtimeStackText = ref('')
const stackFilterMinWaitMinutes = ref(0)

const pprofConfig = ref({
  cpuSeconds: 30,
})

const pprofLoading = reactive({
  cpu: false,
  heap: false,
  goroutine: false,
  allocs: false,
  block: false,
  mutex: false,
})

const pprofResults = ref({
  cpu: false,
  cpuText: '',
  heap: false,
  heapText: '',
  heapStats: null as { alloc: number, objects: number } | null,
  goroutine: '',
  goroutineCount: 0,
  allocs: false,
  allocsText: '',
  block: false,
  blockText: '',
  mutex: false,
  mutexText: '',
})

const hasAnyPprofResult = computed(() => {
  const r = pprofResults.value
  return r.cpu || r.heap || !!r.goroutine || r.allocs || r.block || r.mutex
})

const switchLoading = reactive({
  allow_register: false,
  allow_delete_account: false,
  smtp_ssl: false,
  geetest_enabled: false,
  email_verify_enabled: false,
  sms_verify_enabled: false,
  payment_enabled: false,
})

const langOptions = [
  { label: '中文简体', value: 'zhCN' },
  { label: 'English', value: 'enUS' },
]

const smsProviderOptions = [
  { label: '控制台日志 (开发)', value: 'console' },
  { label: '阿里云短信', value: 'aliyun' },
  { label: '腾讯云短信', value: 'tencent' },
]

const typeOptions = [
  { label: '字符串', value: 'string' },
  { label: '数字', value: 'number' },
  { label: '布尔值', value: 'boolean' },
  { label: 'JSON', value: 'json' },
]

const basicForm = reactive({
  site_name: '',
  site_desc: '',
  site_logo: '',
  copyright: '',
  icp: '',
  version: '',
  default_lang: 'zhCN',
  allow_register: true,
  allow_delete_account: false,
  frontend_url: '',
  backend_api_url: '',
})

const emailForm = reactive({
  email_verify_enabled: true,
  smtp_host: '',
  smtp_port: 587,
  smtp_username: '',
  smtp_password: '',
  smtp_ssl: true,
  system_email_name: '',
})

const smsForm = reactive({
  sms_verify_enabled: false,
  sms_provider: 'console',
  sms_access_key: '',
  sms_secret_key: '',
  sms_sign_name: '',
  sms_template_code: '',
  sms_region: '',
})

const savingSms = ref(false)

const securityForm = reactive({
  geetest_enabled: false,
  geetest_captcha_id: '',
  geetest_captcha_key: '',
  jwt_access_expire: 7200,
  jwt_refresh_expire: 604800,
  login_max_failure: 5,
  login_lock_duration: 10,
  allow_delete_account: false,
})

const paymentForm = reactive({
  payment_enabled: false,
  payment_order_expire_minutes: 30,
})

const customSettings = ref<SettingDTO[]>([])

const addFormRef = ref()
const addForm = reactive({
  key: '',
  value: '',
  label: '',
  type: 'string' as string,
  description: '',
  is_public: false,
})

const addFormRules = {
  key: [
    { required: true, message: '请输入配置键名', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9_]*$/, message: '键名必须以小写字母开头，只能包含小写字母、数字和下划线', trigger: 'blur' },
  ],
  label: [{ required: true, message: '请输入显示名称', trigger: 'blur' }],
}

const editForm = reactive({
  key: '',
  value: '',
  label: '',
  type: 'string' as SettingType,
  description: '',
  is_public: false,
})

const plugins = ref([
  { name: 'Demo Plugin', version: '1.0.0', active: true },
  { name: 'Email', version: '1.0.0', active: true },
])

const customColumns: DataTableColumns<SettingDTO> = [
  { title: '键名', key: 'key' },
  { title: '显示名称', key: 'label' },
  { title: '值', key: 'value', ellipsis: { tooltip: true } },
  { title: '类型', key: 'type', width: 80 },
  {
    title: '公开',
    key: 'is_public',
    width: 80,
    render: row => row.is_public ? '是' : '否',
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render: (row) => {
      return h(NSpace, { size: 8 }, {
        default: () => [
          h(NButton, {
            size: 'small',
            type: 'primary',
            text: true,
            onClick: () => handleEditSetting(row),
          }, () => '修改'),
          h(NButton, {
            size: 'small',
            type: 'error',
            text: true,
            onClick: () => handleDeleteSetting(row.key),
          }, () => '删除'),
        ],
      })
    },
  },
]

type ServiceStatusRow = {
  name: string
  status: 'up' | 'down' | 'warning'
  message: string
  detail: string
}

const serviceStatusRows = computed<ServiceStatusRow[]>(() => {
  if (!serverMonitoringData.value?.services) {
    return []
  }
  return serverMonitoringData.value.services.map(service => {
    const detailParts: string[] = []
    if (typeof service.open_connections === 'number') {
      detailParts.push(`连接:${service.open_connections}`)
    }
    if (typeof service.in_use === 'number') {
      detailParts.push(`使用中:${service.in_use}`)
    }
    if (typeof service.idle === 'number') {
      detailParts.push(`空闲:${service.idle}`)
    }
    if (service.host && service.port) {
      detailParts.push(`${service.host}:${service.port}`)
    }
    return {
      name: service.name,
      status: service.status,
      message: service.message,
      detail: detailParts.join(' | '),
    }
  })
})

const serviceStatusColumns: DataTableColumns<ServiceStatusRow> = [
  { title: '服务', key: 'name' },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: row => {
      const type = row.status === 'up' ? 'success' : row.status === 'warning' ? 'warning' : 'error'
      const text = row.status === 'up' ? '正常' : row.status === 'warning' ? '未就绪' : '异常'
      return h(NTag, { type, size: 'small' }, () => text)
    },
  },
  { title: '说明', key: 'message' },
  { title: '详情', key: 'detail' },
]

const cpuPercent = computed(() => normalizePercent(serverMonitoringData.value?.metrics.cpu.usage_percent ?? 0))
const memoryPercent = computed(() => normalizePercent(serverMonitoringData.value?.metrics.memory.used_percent ?? 0))
const swapPercent = computed(() => normalizePercent(serverMonitoringData.value?.metrics.swap.used_percent ?? 0))
const diskPercent = computed(() => normalizePercent(serverMonitoringData.value?.metrics.disk.used_percent ?? 0))

function normalizePercent(value: number): number {
  if (!Number.isFinite(value)) {
    return 0
  }
  if (value < 0) {
    return 0
  }
  if (value > 100) {
    return 100
  }
  return Number(value.toFixed(2))
}

function formatPercent(value: number): string {
  return `${normalizePercent(value).toFixed(2)}%`
}

function formatInteger(value: number): string {
  if (!Number.isFinite(value)) {
    return '-'
  }
  return Math.round(value).toLocaleString()
}

function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return '-'
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = value
  let idx = 0
  while (size >= 1024 && idx < units.length - 1) {
    size /= 1024
    idx++
  }
  return `${size.toFixed(2)} ${units[idx]}`
}

function formatStorageFromMB(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return '-'
  }
  const gb = value / 1024
  if (gb >= 1024) {
    return `${(gb / 1024).toFixed(2)} TB`
  }
  if (gb >= 1) {
    return `${gb.toFixed(2)} GB`
  }
  return `${value.toFixed(2)} MB`
}

function formatStorageFromGB(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return '-'
  }
  if (value >= 1024) {
    return `${(value / 1024).toFixed(2)} TB`
  }
  return `${value.toFixed(2)} GB`
}

function formatGeneratedAt(value: string | undefined) {
  if (!value) {
    return ''
  }
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) {
    return value
  }
  return d.toLocaleString()
}

function formatUptime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) {
    return '-'
  }
  const day = Math.floor(seconds / 86400)
  const hour = Math.floor((seconds % 86400) / 3600)
  const minute = Math.floor((seconds % 3600) / 60)
  return `${day}天 ${hour}小时 ${minute}分钟`
}

function formatUptimePrecise(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) {
    return '-'
  }
  const day = Math.floor(seconds / 86400)
  const hour = Math.floor((seconds % 86400) / 3600)
  const minute = Math.floor((seconds % 3600) / 60)
  const second = Math.floor(seconds % 60)
  return `${day}天${hour}时${minute}分${second}秒`
}

function formatStartTimeFromUptime(generatedAt?: string, uptimeSeconds?: number): string {
  if (!generatedAt || !Number.isFinite(uptimeSeconds || NaN)) {
    return '-'
  }
  const generated = new Date(generatedAt)
  if (Number.isNaN(generated.getTime())) {
    return '-'
  }
  const start = new Date(generated.getTime() - (uptimeSeconds || 0) * 1000)
  const mm = String(start.getMonth() + 1).padStart(2, '0')
  const dd = String(start.getDate()).padStart(2, '0')
  const hh = String(start.getHours()).padStart(2, '0')
  const mi = String(start.getMinutes()).padStart(2, '0')
  return `${mm}-${dd} ${hh}:${mi}`
}

const uptimeText = computed(() => formatUptime(serverMonitoringData.value?.uptime_seconds ?? 0))
const uptimeTextPrecise = computed(() => formatUptimePrecise(serverMonitoringData.value?.uptime_seconds ?? 0))
const startTimeText = computed(() => formatStartTimeFromUptime(serverMonitoringData.value?.generated_at, serverMonitoringData.value?.uptime_seconds))

async function loadServerMonitoringStatus() {
  loadingServerMonitoring.value = true
  try {
    const response = await adminApi.settings.serverMonitoring()
    serverMonitoringData.value = response.data ?? null
    serverMonitoringGeneratedAt.value = formatGeneratedAt(response.data?.generated_at)
  }
  catch (error: any) {
    serverMonitoringData.value = null
    serverMonitoringGeneratedAt.value = ''
    message.error('加载服务器监控失败: ' + (error.message || '未知错误'))
  }
  finally {
    loadingServerMonitoring.value = false
  }
}

async function loadSettings() {
  loading.value = true
  try {
    const response = await adminApi.settings.list()
    if (response.data?.categories) {
      for (const category of response.data.categories) {
        for (const item of category.items) {
          if (item.key === 'site_name') basicForm.site_name = String(item.value || '')
          if (item.key === 'site_desc') basicForm.site_desc = String(item.value || '')
          if (item.key === 'site_logo') basicForm.site_logo = String(item.value || '')
          if (item.key === 'copyright') basicForm.copyright = String(item.value || '')
          if (item.key === 'icp') basicForm.icp = String(item.value || '')
          if (item.key === 'version') basicForm.version = String(item.value || '')
          if (item.key === 'default_lang') basicForm.default_lang = String(item.value || 'zhCN')
          if (item.key === 'allow_register') basicForm.allow_register = Boolean(item.value)
          if (item.key === 'allow_delete_account') securityForm.allow_delete_account = Boolean(item.value)
          if (item.key === 'frontend_url') basicForm.frontend_url = String(item.value || '')
          if (item.key === 'backend_api_url') basicForm.backend_api_url = String(item.value || '')

          if (item.key === 'email_verify_enabled') emailForm.email_verify_enabled = Boolean(item.value)
          if (item.key === 'smtp_host') emailForm.smtp_host = String(item.value || '')
          if (item.key === 'smtp_port') emailForm.smtp_port = Number(item.value) || 587
          if (item.key === 'smtp_username') emailForm.smtp_username = String(item.value || '')
          if (item.key === 'smtp_password') emailForm.smtp_password = String(item.value || '')
          if (item.key === 'smtp_ssl') emailForm.smtp_ssl = Boolean(item.value)
          if (item.key === 'system_email_name') emailForm.system_email_name = String(item.value || '')

          if (item.key === 'sms_verify_enabled') smsForm.sms_verify_enabled = Boolean(item.value)
          if (item.key === 'sms_provider') smsForm.sms_provider = String(item.value || 'console')
          if (item.key === 'sms_access_key') smsForm.sms_access_key = String(item.value || '')
          if (item.key === 'sms_secret_key') smsForm.sms_secret_key = String(item.value || '')
          if (item.key === 'sms_sign_name') smsForm.sms_sign_name = String(item.value || '')
          if (item.key === 'sms_template_code') smsForm.sms_template_code = String(item.value || '')
          if (item.key === 'sms_region') smsForm.sms_region = String(item.value || '')

          if (item.key === 'geetest_enabled') securityForm.geetest_enabled = Boolean(item.value)
          if (item.key === 'geetest_captcha_id') securityForm.geetest_captcha_id = String(item.value || '')
          if (item.key === 'geetest_captcha_key') securityForm.geetest_captcha_key = String(item.value || '')
          if (item.key === 'jwt_access_expire') securityForm.jwt_access_expire = Number(item.value) || 7200
          if (item.key === 'jwt_refresh_expire') securityForm.jwt_refresh_expire = Number(item.value) || 604800
          if (item.key === 'login_max_failure') securityForm.login_max_failure = Number(item.value) || 5
          if (item.key === 'login_lock_duration') securityForm.login_lock_duration = Number(item.value) || 10

          if (item.key === 'payment_enabled') paymentForm.payment_enabled = Boolean(item.value)
          if (item.key === 'payment_order_expire_minutes') paymentForm.payment_order_expire_minutes = Number(item.value) || 30
        }

        if (category.category === 'custom') {
          customSettings.value = category.items
        }
      }
    }
  }
  catch (error: any) {
    message.error('加载配置失败: ' + (error.message || '未知错误'))
  }
  finally {
    loading.value = false
  }
}

async function handleUpdateAllowRegister(nextValue: boolean) {
  const prev = basicForm.allow_register
  basicForm.allow_register = nextValue
  switchLoading.allow_register = true
  try {
    await adminApi.settings.update('allow_register', String(nextValue))
    settingsStore.updateConfig({ allow_register: nextValue })
    message.success('注册开关已更新')
  }
  catch (error: any) {
    basicForm.allow_register = prev
    message.error('更新失败: ' + (error.message || '未知错误'))
  }
  finally {
    switchLoading.allow_register = false
  }
}

async function handleUpdateAllowDeleteAccount(nextValue: boolean) {
  const prev = securityForm.allow_delete_account
  securityForm.allow_delete_account = nextValue
  switchLoading.allow_delete_account = true
  try {
    await adminApi.settings.update('allow_delete_account', String(nextValue))
    settingsStore.updateConfig({ allow_delete_account: nextValue })
    message.success('账号注销开关已更新')
  }
  catch (error: any) {
    securityForm.allow_delete_account = prev
    message.error('更新失败: ' + (error.message || '未知错误'))
  }
  finally {
    switchLoading.allow_delete_account = false
  }
}

async function handleUpdateSmtpSSL(nextValue: boolean) {
  const prev = emailForm.smtp_ssl
  emailForm.smtp_ssl = nextValue
  switchLoading.smtp_ssl = true
  try {
    await adminApi.settings.update('smtp_ssl', String(nextValue))
    message.success('SMTP SSL 开关已更新')
  }
  catch (error: any) {
    emailForm.smtp_ssl = prev
    message.error('更新失败: ' + (error.message || '未知错误'))
  }
  finally {
    switchLoading.smtp_ssl = false
  }
}

async function handleUpdateEmailVerifyEnabled(nextValue: boolean) {
  const prev = emailForm.email_verify_enabled
  emailForm.email_verify_enabled = nextValue
  switchLoading.email_verify_enabled = true
  try {
    await adminApi.settings.update('email_verify_enabled', String(nextValue))
    settingsStore.updateConfig({ email_verify_enabled: nextValue })
    message.success('邮箱验证码开关已更新')
  }
  catch (error: any) {
    emailForm.email_verify_enabled = prev
    message.error('更新失败: ' + (error.message || '未知错误'))
  }
  finally {
    switchLoading.email_verify_enabled = false
  }
}

async function handleUpdateSmsVerifyEnabled(nextValue: boolean) {
  const prev = smsForm.sms_verify_enabled
  smsForm.sms_verify_enabled = nextValue
  switchLoading.sms_verify_enabled = true
  try {
    await adminApi.settings.update('sms_verify_enabled', String(nextValue))
    settingsStore.updateConfig({ sms_verify_enabled: nextValue })
    message.success('短信验证码开关已更新')
  }
  catch (error: any) {
    smsForm.sms_verify_enabled = prev
    message.error('更新失败: ' + (error.message || '未知错误'))
  }
  finally {
    switchLoading.sms_verify_enabled = false
  }
}

async function handleSaveSms() {
  savingSms.value = true
  try {
    await adminApi.settings.batchUpdate({
      sms_provider: smsForm.sms_provider,
      sms_access_key: smsForm.sms_access_key,
      sms_secret_key: smsForm.sms_secret_key,
      sms_sign_name: smsForm.sms_sign_name,
      sms_template_code: smsForm.sms_template_code,
      sms_region: smsForm.sms_region,
    })
    message.success('短信设置保存成功')
  }
  catch (error: any) {
    message.error('保存失败: ' + (error.message || '未知错误'))
  }
  finally {
    savingSms.value = false
  }
}

async function handleUpdateGeetestEnabled(nextValue: boolean) {
  const prev = securityForm.geetest_enabled
  securityForm.geetest_enabled = nextValue
  switchLoading.geetest_enabled = true
  try {
    await adminApi.settings.update('geetest_enabled', String(nextValue))
    settingsStore.updateConfig({ geetest_enabled: nextValue })
    message.success('极验开关已更新')
  }
  catch (error: any) {
    securityForm.geetest_enabled = prev
    message.error('更新失败: ' + (error.message || '未知错误'))
  }
  finally {
    switchLoading.geetest_enabled = false
  }
}

async function handleUpdatePaymentEnabled(nextValue: boolean) {
  const prev = paymentForm.payment_enabled
  paymentForm.payment_enabled = nextValue
  switchLoading.payment_enabled = true
  try {
    await adminApi.settings.update('payment_enabled', String(nextValue))
    message.success('支付功能开关已更新')
  }
  catch (error: any) {
    paymentForm.payment_enabled = prev
    message.error('更新失败: ' + (error.message || '未知错误'))
  }
  finally {
    switchLoading.payment_enabled = false
  }
}

async function handleSavePayment() {
  savingPayment.value = true
  try {
    await adminApi.settings.batchUpdate({
      payment_order_expire_minutes: String(paymentForm.payment_order_expire_minutes),
    })
    message.success('支付设置保存成功')
  }
  catch (error: any) {
    message.error('保存失败: ' + (error.message || '未知错误'))
  }
  finally {
    savingPayment.value = false
  }
}

async function handleSaveBasic() {
  savingBasic.value = true
  try {
    const frontendUrl = basicForm.frontend_url.trim().replace(/\/+$/, '')
    const backendApiUrl = basicForm.backend_api_url.trim().replace(/\/+$/, '')
    basicForm.frontend_url = frontendUrl
    basicForm.backend_api_url = backendApiUrl
    await adminApi.settings.batchUpdate({
      site_name: basicForm.site_name,
      site_desc: basicForm.site_desc,
      site_logo: basicForm.site_logo,
      copyright: basicForm.copyright,
      icp: basicForm.icp,
      version: basicForm.version,
      default_lang: basicForm.default_lang,
      frontend_url: frontendUrl,
      backend_api_url: backendApiUrl,
    })
    settingsStore.updateConfig({
      site_name: basicForm.site_name,
      site_desc: basicForm.site_desc,
      site_logo: basicForm.site_logo,
      copyright: basicForm.copyright,
      icp: basicForm.icp,
      version: basicForm.version,
      default_lang: basicForm.default_lang,
    })
    message.success('基本设置保存成功')
  }
  catch (error: any) {
    message.error('保存失败: ' + (error.message || '未知错误'))
  }
  finally {
    savingBasic.value = false
  }
}

async function handleSaveEmail() {
  savingEmail.value = true
  try {
    await adminApi.settings.batchUpdate({
      smtp_host: emailForm.smtp_host,
      smtp_port: String(emailForm.smtp_port),
      smtp_username: emailForm.smtp_username,
      smtp_password: emailForm.smtp_password,
      system_email_name: emailForm.system_email_name,
    })
    message.success('邮件设置保存成功')
  }
  catch (error: any) {
    message.error('保存失败: ' + (error.message || '未知错误'))
  }
  finally {
    savingEmail.value = false
  }
}

async function handleTestEmail() {
  testingEmail.value = true
  try {
    await new Promise(r => setTimeout(r, 1000))
    message.info('测试邮件发送功能开发中...')
  }
  finally {
    testingEmail.value = false
  }
}

async function handleSaveSecurity() {
  savingSecurity.value = true
  try {
    await adminApi.settings.batchUpdate({
      geetest_captcha_id: securityForm.geetest_captcha_id,
      geetest_captcha_key: securityForm.geetest_captcha_key,
      jwt_access_expire: String(securityForm.jwt_access_expire),
      jwt_refresh_expire: String(securityForm.jwt_refresh_expire),
      login_max_failure: String(securityForm.login_max_failure),
      login_lock_duration: String(securityForm.login_lock_duration),
    })
    settingsStore.updateConfig({
      geetest_captcha_id: securityForm.geetest_captcha_id,
    })
    message.success('安全设置保存成功')
  }
  catch (error: any) {
    message.error('保存失败: ' + (error.message || '未知错误'))
  }
  finally {
    savingSecurity.value = false
  }
}

async function handleRestartBackend() {
  restartingBackend.value = true
  try {
    await adminApi.settings.restartBackend()
    message.success('后端重启请求已发送')
  }
  catch (error: any) {
    message.error('重启失败: ' + (error.message || '未知错误'))
  }
  finally {
    restartingBackend.value = false
  }
}

async function handleAddSetting() {
  try {
    await addFormRef.value?.validate()
  }
  catch {
    return
  }

  adding.value = true
  try {
    await adminApi.settings.create({
      key: addForm.key,
      value: addForm.value,
      type: addForm.type as SettingType,
      category: 'custom',
      label: addForm.label,
      description: addForm.description,
      is_public: addForm.is_public,
      is_editable: true,
    })
    message.success('配置项添加成功')
    showAddModal.value = false
    addForm.key = ''
    addForm.value = ''
    addForm.label = ''
    addForm.type = 'string'
    addForm.description = ''
    addForm.is_public = false
    await loadSettings()
  }
  catch (error: any) {
    message.error('添加失败: ' + (error.message || '未知错误'))
  }
  finally {
    adding.value = false
  }
}

async function handleDeleteSetting(key: string) {
  try {
    await adminApi.settings.delete(key)
    message.success('配置项已删除')
    await loadSettings()
  }
  catch (error: any) {
    message.error('删除失败: ' + (error.message || '未知错误'))
  }
}

function handleEditSetting(row: SettingDTO) {
  editForm.key = row.key
  editForm.value = row.value == null ? '' : String(row.value)
  editForm.label = row.label || ''
  editForm.type = row.type
  editForm.description = row.description || ''
  editForm.is_public = Boolean(row.is_public)
  showEditModal.value = true
}

async function handleSaveSettingEdit() {
  if (!editForm.key) {
    return
  }
  savingEdit.value = true
  try {
    await adminApi.settings.updateMeta(editForm.key, {
      value: editForm.value,
      type: editForm.type,
      category: 'custom',
      label: editForm.label,
      description: editForm.description,
      is_public: editForm.is_public,
      is_editable: true,
    })
    message.success('配置项修改成功')
    showEditModal.value = false
    await loadSettings()
  }
  catch (error: any) {
    message.error('修改失败: ' + (error.message || '未知错误'))
  }
  finally {
    savingEdit.value = false
  }
}

async function loadDebugStats() {
  loadingDebugStats.value = true
  try {
    const res = await adminDebugApi.goroutineStats()
    if (res.data) {
      debugStats.value = res.data
    }
  }
  catch (e: any) {
    message.error('加载调试统计失败: ' + e.message)
  }
  finally {
    loadingDebugStats.value = false
  }
}

function toggleDebugAutoRefresh(value: boolean) {
  debugAutoRefresh.value = value
  if (value) {
    debugRefreshInterval.value = window.setInterval(loadDebugStats, 3000)
  }
  else {
    if (debugRefreshInterval.value) {
      clearInterval(debugRefreshInterval.value)
      debugRefreshInterval.value = null
    }
  }
}

async function handleForceGC() {
  try {
    const res = await adminDebugApi.forceGC()
    if (res.data) {
      message.success(`GC完成: ${res.data.goroutines_before} -> ${res.data.goroutines_after} 协程`)
      loadDebugStats()
    }
  }
  catch (e: any) {
    message.error('操作失败: ' + e.message)
  }
}

function clearAllPprofResults() {
  pprofResults.value = {
    cpu: false,
    cpuText: '',
    heap: false,
    heapText: '',
    heapStats: null,
    goroutine: '',
    goroutineCount: 0,
    allocs: false,
    allocsText: '',
    block: false,
    blockText: '',
    mutex: false,
    mutexText: '',
  }
  message.success('已清空所有结果')
}

async function captureCPUProfile() {
  pprofLoading.cpu = true
  pprofResults.value.cpu = false
  pprofResults.value.cpuText = ''
  message.info(`开始采集 CPU Profile (${pprofConfig.value.cpuSeconds}秒)...`)
  try {
    const url = adminDebugApi.cpuProfile(pprofConfig.value.cpuSeconds)
    const token = local.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.cpu = true
      pprofResults.value.cpuText = text
      message.success('CPU Profile 采集完成')
    }
    else {
      message.error('采集失败')
    }
  }
  catch (e: any) {
    message.error('采集失败: ' + e.message)
  }
  finally {
    pprofLoading.cpu = false
  }
}

async function captureHeapProfile() {
  pprofLoading.heap = true
  pprofResults.value.heap = false
  pprofResults.value.heapText = ''
  pprofResults.value.heapStats = null
  try {
    const url = adminDebugApi.heapProfile()
    const token = local.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.heap = true
      pprofResults.value.heapText = text
      const allocMatch = text.match(/# Alloc = (\d+)/)
      const objectsMatch = text.match(/# HeapObjects = (\d+)/)
      if (allocMatch || objectsMatch) {
        pprofResults.value.heapStats = {
          alloc: allocMatch ? Number(allocMatch[1]) : 0,
          objects: objectsMatch ? Number(objectsMatch[1]) : 0,
        }
      }
      message.success('Heap Profile 采集完成')
    }
    else {
      message.error('采集失败')
    }
  }
  catch (e: any) {
    message.error('采集失败: ' + e.message)
  }
  finally {
    pprofLoading.heap = false
  }
}

async function captureGoroutineProfile() {
  pprofLoading.goroutine = true
  pprofResults.value.goroutine = ''
  pprofResults.value.goroutineCount = 0
  try {
    const url = adminDebugApi.goroutineProfile(0)
    const token = local.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.goroutine = text
      const matches = text.match(/goroutine \d+/g)
      pprofResults.value.goroutineCount = matches ? matches.length : 0
      message.success('Goroutine Profile 采集完成')
    }
    else {
      message.error('采集失败')
    }
  }
  catch (e: any) {
    message.error('采集失败: ' + e.message)
  }
  finally {
    pprofLoading.goroutine = false
  }
}

async function captureAllocsProfile() {
  pprofLoading.allocs = true
  pprofResults.value.allocs = false
  pprofResults.value.allocsText = ''
  try {
    const url = adminDebugApi.allocsProfile()
    const token = local.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.allocs = true
      pprofResults.value.allocsText = text
      message.success('Allocs Profile 采集完成')
    }
    else {
      message.error('采集失败')
    }
  }
  catch (e: any) {
    message.error('采集失败: ' + e.message)
  }
  finally {
    pprofLoading.allocs = false
  }
}

async function captureBlockProfile() {
  pprofLoading.block = true
  pprofResults.value.block = false
  pprofResults.value.blockText = ''
  try {
    const url = adminDebugApi.blockProfile()
    const token = local.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.block = true
      pprofResults.value.blockText = text
      message.success('Block Profile 采集完成')
    }
    else {
      message.error('采集失败')
    }
  }
  catch (e: any) {
    message.error('采集失败: ' + e.message)
  }
  finally {
    pprofLoading.block = false
  }
}

async function captureMutexProfile() {
  pprofLoading.mutex = true
  pprofResults.value.mutex = false
  pprofResults.value.mutexText = ''
  try {
    const url = adminDebugApi.mutexProfile()
    const token = local.get('accessToken')
    const res = await fetch(url, {
      headers: { 'Authorization': `Bearer ${token}` },
    })
    if (res.ok) {
      const text = await res.text()
      pprofResults.value.mutex = true
      pprofResults.value.mutexText = text
      message.success('Mutex Profile 采集完成')
    }
    else {
      message.error('采集失败')
    }
  }
  catch (e: any) {
    message.error('采集失败: ' + e.message)
  }
  finally {
    pprofLoading.mutex = false
  }
}

async function loadRuntimeStacks() {
  loadingRuntimeStacks.value = true
  runtimeStackText.value = ''
  try {
    const url = adminDebugApi.goroutineProfile(stackFilterMinWaitMinutes.value)
    const token = local.get('accessToken')
    const res = await fetch(url, {
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    })
    if (res.ok) {
      const text = await res.text()
      runtimeStackText.value = text
      const filterMsg = stackFilterMinWaitMinutes.value > 0 ? `（已过滤 >${stackFilterMinWaitMinutes.value}分钟）` : ''
      message.success(`堆栈加载完成${filterMsg}`)
    }
    else {
      message.error('加载失败')
    }
  }
  catch (e: any) {
    message.error('加载失败: ' + e.message)
  }
  finally {
    loadingRuntimeStacks.value = false
  }
}

function clearRuntimeStacks() {
  runtimeStackText.value = ''
  message.success('已清空堆栈')
}

onMounted(() => {
  loadSettings()
  loadServerMonitoringStatus()
})

onUnmounted(() => {
  if (debugRefreshInterval.value) {
    clearInterval(debugRefreshInterval.value)
  }
})
</script>

<style scoped></style>
