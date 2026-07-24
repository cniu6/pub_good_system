<script setup lang="ts">
import TableColumnSelector from '@/components/common/TableColumnSelector.vue'
import {
  CheckCircleOutlined,
  ReloadOutlined,
} from '@vicons/antd'
import { useAdminDashboard } from './composables/useAdminDashboard'

const {
  t,
  mode,
  loading,
  alertOnlyIssues,
  lastRefreshAt,
  statistics,
  recentUsers,
  recentLoginUsers,
  dashboardFailedMetrics,
  monitoring,
  operations,
  topSummaryCards,
  quick_actions,
  resourceRows,
  serviceItems,
  operationsSummary,
  taskHighlights,
  rateLimitHighlights,
  displayedAlertItems,
  alertSeveritySummary,
  alertCountTagType,
  systemShortcutActions,
  recentUserColumnOptions,
  recentUserSelectedColumnKeys,
  recentUserVisibleColumns,
  recentUserVisibleColumnCount,
  recentUserTotalColumnCount,
  recentUserTableScrollX,
  resetRecentUserSelectedColumns,
  recentLoginColumnOptions,
  recentLoginSelectedColumnKeys,
  recentLoginVisibleColumns,
  recentLoginVisibleColumnCount,
  recentLoginTotalColumnCount,
  recentLoginTableScrollX,
  resetRecentLoginSelectedColumns,
  formatCurrency,
  formatDateTime,
  formatBytes,
  formatUptime,
  getServiceTagType,
  getServiceStatusText,
  formatServiceMeta,
  getTaskTagType,
  getTaskStatusText,
  formatTaskMeta,
  formatTaskDuration,
  getRateLimitTagType,
  formatRateLimitMeta,
  getAlertLevelLabel,
  handleAlertClick,
  handleShortcutAction,
  handleRefresh,
  go_to,
} = useAdminDashboard()
</script>

<template>
  <n-space vertical :size="16">
    <!-- 欢迎横幅 -->
    <n-card hoverable>
      <n-flex justify="space-between" align="center" wrap :size="16">
        <n-flex align="center" :size="16">
          <n-icon-wrapper :size="48" :border-radius="12" color="var(--success-color)">
            <n-icon :size="26" color="#fff">
              <CheckCircleOutlined />
            </n-icon>
          </n-icon-wrapper>
          <n-flex vertical>
            <n-text strong>
              {{ t('adminDashboard.welcomeBack') }}
            </n-text>
            <n-text depth="3">
              {{ t('adminDashboard.systemOverview') }}
            </n-text>
          </n-flex>
        </n-flex>
        <n-flex :size="8">
          <n-button :loading="loading" @click="handleRefresh">
            {{ t('adminDashboard.refreshData') }}
          </n-button>
          <n-button type="primary" @click="go_to('finance/payment-orders')">
            {{ t('route.paymentOrders') }}
          </n-button>
          <n-button @click="go_to('settings/server-management')">
            {{ t('route.serverManagement') }}
          </n-button>
        </n-flex>
      </n-flex>
    </n-card>

    <!-- 部分统计指标查库失败时的提示：数值仍会用 0 兜底展示，避免误以为「真的没有数据」 -->
    <n-alert v-if="dashboardFailedMetrics.length > 0" type="warning" closable>
      {{ t('adminDashboard.partialDataWarning', { metrics: dashboardFailedMetrics.join(', ') }) }}
    </n-alert>

    <n-grid :x-gap="16" :y-gap="16" :cols="3" item-responsive responsive="screen">
      <n-gi v-for="card in topSummaryCards" :key="card.key" span="3 m:1">
        <n-card hoverable class="dashboard-summary-card">
          <n-space vertical :size="16">
            <n-flex align="center" :size="12" class="dashboard-summary-header">
              <n-icon-wrapper :size="42" :color="card.color" :border-radius="12">
                <n-icon :size="22" color="#fff">
                  <component :is="card.icon" />
                </n-icon>
              </n-icon-wrapper>
              <n-text strong>
                {{ card.title }}
              </n-text>
            </n-flex>

            <n-space vertical :size="12">
              <n-flex v-for="metric in card.metrics" :key="metric.key" justify="space-between" align="center" class="dashboard-summary-row">
                <n-text depth="3">
                  {{ metric.label }}
                </n-text>
                <div class="dashboard-summary-value">
                  <span v-if="metric.prefix">{{ metric.prefix }}</span>
                  <n-number-animation :from="0" :to="metric.value" :precision="metric.precision" show-separator />
                  <span v-if="metric.suffix">{{ metric.suffix }}</span>
                </div>
              </n-flex>
            </n-space>
          </n-space>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 下半区域 -->
    <n-grid :x-gap="16" :y-gap="16" :cols="12" item-responsive responsive="screen">
      <n-gi span="12 xl:8">
        <n-card :title="t('adminDashboard.businessTrend')" hoverable>
          <div ref="businessTrendRef" style="height: 320px;" />
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminDashboard.verifyDistribution')" hoverable>
          <template #header-extra>
            <n-text depth="3">
              {{ t('adminRealname.totalRecords', { total: statistics.total_realname_requests }) }}
            </n-text>
          </template>
          <div ref="verifyTrendRef" style="height: 320px;" />
        </n-card>
      </n-gi>

      <n-gi span="12 xl:8">
        <n-card :title="t('adminDashboard.logHistoryTrend')" hoverable>
          <div ref="logTrendRef" style="height: 320px;" />
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminSettings.processResources')" hoverable>
          <n-space vertical :size="14">
            <div v-for="item in resourceRows" :key="item.label">
              <n-flex justify="space-between" align="center">
                <n-text>{{ item.label }}</n-text>
                <n-text depth="3">
                  {{ item.detail }}
                </n-text>
              </n-flex>
              <n-progress type="line" :percentage="item.percentage" :status="item.percentage >= 85 ? 'error' : item.percentage >= 70 ? 'warning' : 'success'" />
            </div>
          </n-space>
        </n-card>
      </n-gi>

      <!-- 运行概览 -->
      <n-gi span="12 xl:4">
        <n-card :title="t('adminDashboard.financeOverview')" hoverable>
          <n-descriptions :column="1" label-placement="left" bordered size="small">
            <n-descriptions-item :label="t('adminDashboard.paidOrders')">
              {{ statistics.paid_payment_orders }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminDashboard.pendingOrders')">
              {{ statistics.pending_payment_orders }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.statApproved')">
              {{ statistics.approved_withdraw_count }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminWithdraw.statPaidCount')">
              {{ statistics.paid_withdraw_count }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminDashboard.paidWithdrawAmount')">
              ¥{{ formatCurrency(statistics.paid_withdraw_amount) }}
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminSettings.systemInfo')" hoverable>
          <template #header-extra>
            <n-text depth="3">
              {{ t('adminSettings.lastRefreshed') }} {{ formatDateTime(monitoring?.generated_at) }}
            </n-text>
          </template>
          <n-descriptions :column="1" label-placement="left" bordered size="small">
            <n-descriptions-item :label="t('adminSettings.appName')">
              {{ monitoring?.app.name || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.appMode')">
              <n-tag size="small" :type="mode === 'production' ? 'success' : 'warning'">
                {{ monitoring?.app.mode || mode }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.port')">
              {{ monitoring?.app.port || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.pid')">
              {{ monitoring?.process.pid || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.goVersion')">
              {{ monitoring?.app.go_version || '-' }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.goroutines')">
              {{ monitoring?.process.goroutines || 0 }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.uptime')">
              {{ formatUptime(monitoring?.uptime_seconds) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.upload')">
              {{ formatBytes(monitoring?.metrics.network.bytes_sent) }}
            </n-descriptions-item>
            <n-descriptions-item :label="t('adminSettings.download')">
              {{ formatBytes(monitoring?.metrics.network.bytes_recv) }}
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminSettings.serviceHealthSnapshot')" hoverable>
          <n-space vertical :size="12">
            <n-card v-for="service in serviceItems" :key="service.name" size="small" embedded>
              <n-space vertical :size="8">
                <n-flex justify="space-between" align="center" wrap>
                  <n-space align="center" :size="8">
                    <n-text strong>
                      {{ service.name }}
                    </n-text>
                    <n-tag size="small" :type="getServiceTagType(service.status)">
                      {{ getServiceStatusText(service.status) }}
                    </n-tag>
                  </n-space>
                </n-flex>
                <n-text depth="3">
                  {{ formatServiceMeta(service) }}
                </n-text>
              </n-space>
            </n-card>
            <n-empty v-if="serviceItems.length === 0 && !loading" />
          </n-space>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:6">
        <n-card :title="t('adminDashboard.operationsOverview')" hoverable>
          <template #header-extra>
            <n-button text type="primary" @click="go_to('settings/server-management')">
              {{ t('route.serverManagement') }}
            </n-button>
          </template>
          <n-space vertical :size="16">
            <n-grid :x-gap="12" :y-gap="12" :cols="2">
              <n-gi v-for="item in operationsSummary" :key="item.label">
                <n-card size="small" embedded>
                  <n-statistic :label="item.label" :value="item.value" />
                </n-card>
              </n-gi>
            </n-grid>

            <n-divider style="margin: 0;" />

            <n-space vertical :size="10">
              <n-text strong>
                {{ t('adminServer.tasks.title') }}
              </n-text>
              <n-card v-for="task in taskHighlights" :key="task.key" size="small" embedded>
                <n-space vertical :size="8">
                  <n-flex justify="space-between" align="center" wrap>
                    <n-space align="center" :size="8">
                      <n-text strong>
                        {{ task.label }}
                      </n-text>
                      <n-tag size="small" :type="getTaskTagType(task)">
                        {{ getTaskStatusText(task) }}
                      </n-tag>
                    </n-space>
                    <n-text depth="3">
                      {{ task.interval_secs }}s
                    </n-text>
                  </n-flex>
                  <n-space :size="12" wrap>
                    <n-text depth="3">
                      {{ t('adminServer.tasks.lastRun') }}: {{ formatDateTime(task.last_run_time) }}
                    </n-text>
                    <n-text depth="3">
                      {{ t('adminServer.tasks.duration') }}: {{ formatTaskDuration(task.last_duration_ms) }}
                    </n-text>
                  </n-space>
                  <n-text depth="3">
                    {{ formatTaskMeta(task) }}
                  </n-text>
                </n-space>
              </n-card>
              <n-empty v-if="taskHighlights.length === 0 && !loading" />
            </n-space>

            <n-divider style="margin: 0;" />

            <n-space vertical :size="10">
              <n-flex justify="space-between" align="center" wrap>
                <n-text strong>
                  {{ t('adminServer.rateLimit.title') }}
                </n-text>
                <n-tag size="small" :type="operations?.api_log.enabled ? 'success' : 'default'">
                  {{ t('adminServer.runtimeConfig.apiLog') }}: {{ operations?.api_log.enabled ? t('common.enable') : t('common.disable') }}
                </n-tag>
              </n-flex>
              <n-card v-for="item in rateLimitHighlights" :key="item.name" size="small" embedded>
                <n-space vertical :size="8">
                  <n-flex justify="space-between" align="center" wrap>
                    <n-space align="center" :size="8">
                      <n-text strong>
                        {{ item.name }}
                      </n-text>
                      <n-tag size="small" :type="getRateLimitTagType(item)">
                        {{ item.enabled ? t('common.enable') : t('common.disable') }}
                      </n-tag>
                    </n-space>
                    <n-text depth="3">
                      R{{ item.rate }} / B{{ item.burst }}
                    </n-text>
                  </n-flex>
                  <n-text depth="3">
                    {{ formatRateLimitMeta(item) }}
                  </n-text>
                </n-space>
              </n-card>
              <n-empty v-if="rateLimitHighlights.length === 0 && !loading" />
            </n-space>
          </n-space>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:6">
        <n-card :title="t('adminDashboard.alertCenter')" hoverable>
          <template #header-extra>
            <n-flex align="center" :size="8" wrap>
              <n-text depth="3">
                {{ t('adminDashboard.lastUpdated') }} {{ formatDateTime(lastRefreshAt) }}
              </n-text>
              <n-tag v-if="alertSeveritySummary.error" size="small" type="error">
                {{ t('adminDashboard.criticalLevel') }} {{ alertSeveritySummary.error }}
              </n-tag>
              <n-tag v-if="alertSeveritySummary.warning" size="small" type="warning">
                {{ t('adminDashboard.warningLevel') }} {{ alertSeveritySummary.warning }}
              </n-tag>
              <n-tag v-if="alertSeveritySummary.info" size="small" type="info">
                {{ t('adminDashboard.infoLevel') }} {{ alertSeveritySummary.info }}
              </n-tag>
              <n-tag size="small" :type="alertCountTagType">
                {{ displayedAlertItems.length }}
              </n-tag>
              <n-text depth="3">
                {{ t('adminDashboard.onlyIssues') }}
              </n-text>
              <n-switch v-model:value="alertOnlyIssues" size="small" />
            </n-flex>
          </template>
          <n-space vertical :size="12">
            <n-alert v-for="item in displayedAlertItems" :key="item.key" :type="item.type" :title="item.title">
              <n-space vertical :size="8">
                <n-tag size="small" :type="item.type === 'info' ? 'info' : item.type">
                  {{ getAlertLevelLabel(item.type) }}
                </n-tag>
                <span>{{ item.detail }}</span>
                <n-button v-if="item.path" size="tiny" tertiary type="primary" @click="handleAlertClick(item)">
                  {{ item.actionLabel || t('adminDashboard.viewDetails') }}
                </n-button>
              </n-space>
            </n-alert>
            <n-empty v-if="displayedAlertItems.length === 0 && !loading" :description="alertOnlyIssues ? t('adminDashboard.noIssueAlerts') : t('adminDashboard.noAlerts')" />
          </n-space>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminDashboard.systemActions')" hoverable>
          <n-space vertical :size="12">
            <n-grid :x-gap="12" :y-gap="12" :cols="2" item-responsive>
              <n-gi v-for="action in systemShortcutActions" :key="action.key" span="2 s:1">
                <n-card size="small" embedded>
                  <n-space vertical :size="10">
                    <n-flex justify="space-between" align="center">
                      <n-text strong>
                        {{ action.label }}
                      </n-text>
                      <n-icon>
                        <component :is="action.icon" />
                      </n-icon>
                    </n-flex>
                    <n-text depth="3">
                      {{ action.description }}
                    </n-text>
                    <n-button block :type="action.type" tertiary @click="handleShortcutAction(action)">
                      {{ action.actionLabel }}
                    </n-button>
                  </n-space>
                </n-card>
              </n-gi>
            </n-grid>
            <n-divider style="margin: 0;" />
            <n-space :size="8" wrap>
              <n-button :loading="loading" @click="handleRefresh">
                <template #icon>
                  <n-icon><ReloadOutlined /></n-icon>
                </template>
                {{ t('adminSettings.refresh') }}
              </n-button>
            </n-space>
          </n-space>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:4">
        <n-card :title="t('adminDashboard.quickActions')" hoverable>
          <n-grid :x-gap="12" :y-gap="12" :cols="2" item-responsive>
            <n-gi v-for="action in quick_actions" :key="action.label" span="2 s:1">
              <n-button block :type="action.type" ghost size="large" @click="go_to(action.path)">
                <template #icon>
                  <n-icon><component :is="action.icon" /></n-icon>
                </template>
                {{ action.label }}
              </n-button>
            </n-gi>
          </n-grid>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:6">
        <n-card :title="t('adminDashboard.recentUsers')" hoverable>
          <template #header-extra>
            <n-space>
              <TableColumnSelector
                v-model="recentUserSelectedColumnKeys"
                :options="recentUserColumnOptions"
                :visible-count="recentUserVisibleColumnCount"
                :total-count="recentUserTotalColumnCount"
                :button-label="t('common.showFields')"
                :title="t('common.visibleFields')"
                :hint="t('common.columnVisibilityHint')"
                :reset-label="t('common.restoreDefaultFields')"
                @reset="resetRecentUserSelectedColumns"
              />
              <n-button type="primary" quaternary @click="go_to('users')">
                {{ t('adminDashboard.viewAll') }}
              </n-button>
            </n-space>
          </template>
          <n-spin :show="loading">
            <n-data-table
              :columns="recentUserVisibleColumns"
              :data="recentUsers"
              :bordered="false"
              :single-line="false"
              size="small"
              :pagination="false"
              :scroll-x="recentUserTableScrollX"
            />
            <n-empty v-if="!loading && recentUsers.length === 0" :description="t('adminDashboard.noUserData')" />
          </n-spin>
        </n-card>
      </n-gi>

      <n-gi span="12 xl:6">
        <n-card :title="t('adminDashboard.recentLoginUsers')" hoverable>
          <template #header-extra>
            <n-space>
              <TableColumnSelector
                v-model="recentLoginSelectedColumnKeys"
                :options="recentLoginColumnOptions"
                :visible-count="recentLoginVisibleColumnCount"
                :total-count="recentLoginTotalColumnCount"
                :button-label="t('common.showFields')"
                :title="t('common.visibleFields')"
                :hint="t('common.columnVisibilityHint')"
                :reset-label="t('common.restoreDefaultFields')"
                @reset="resetRecentLoginSelectedColumns"
              />
              <n-button type="primary" quaternary @click="go_to('users')">
                {{ t('adminDashboard.viewAll') }}
              </n-button>
            </n-space>
          </template>
          <n-spin :show="loading">
            <n-data-table
              :columns="recentLoginVisibleColumns"
              :data="recentLoginUsers"
              :bordered="false"
              :single-line="false"
              size="small"
              :pagination="false"
              :scroll-x="recentLoginTableScrollX"
            />
            <n-empty v-if="!loading && recentLoginUsers.length === 0" :description="t('adminDashboard.noUserData')" />
          </n-spin>
        </n-card>
      </n-gi>
    </n-grid>
  </n-space>
</template>

<style scoped>
.dashboard-summary-card {
  height: 100%;
}

.dashboard-summary-header {
  min-height: 42px;
}

.dashboard-summary-row {
  gap: 12px;
}

.dashboard-summary-value {
  display: flex;
  align-items: baseline;
  gap: 2px;
  font-size: 18px;
  font-weight: 600;
  line-height: 1;
  white-space: nowrap;
}
</style>
