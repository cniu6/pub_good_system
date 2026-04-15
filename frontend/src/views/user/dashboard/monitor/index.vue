<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/store'
import Chart from './components/chart.vue'
import Chart2 from './components/chart2.vue'
import Chart3 from './components/chart3.vue'

const appStore = useAppStore()
const { t } = useI18n()

const tableData = [
  {
    id: 0,
    name: t('monitor.productName1'),
    start: '2022-02-02',
    end: '2022-02-02',
    prograss: '100',
    status: t('monitor.completed'),
  },
  {
    id: 0,
    name: t('monitor.productName2'),
    start: '2022-02-02',
    end: '2022-02-02',
    prograss: '50',
    status: t('monitor.inProgress'),
  },
  {
    id: 0,
    name: t('monitor.productName3'),
    start: '2022-02-02',
    end: '2022-02-02',
    prograss: '100',
    status: t('monitor.completed'),
  },
]
</script>

<template>
  <n-grid
    :x-gap="16"
    :y-gap="16"
    :cols="12"
    item-responsive
    responsive="screen"
  >
    <!-- 统计卡片 - 移动端每行2个，桌面端每行4个 -->
    <n-gi span="6 m:3">
      <n-card>
        <n-space
          justify="space-between"
          align="center"
        >
          <n-statistic :label="t('monitor.visits')">
            <n-number-animation
              :from="0"
              :to="12039"
              show-separator
            />
          </n-statistic>
          <n-icon
            color="#de4307"
            size="42"
          >
            <icon-park-outline-chart-histogram />
          </n-icon>
        </n-space>
        <template #footer>
          <n-space justify="space-between">
            <span>{{ t('monitor.totalVisits') }}</span>
            <span><n-number-animation
              :from="0"
              :to="322039"
              show-separator
            /></span>
          </n-space>
        </template>
      </n-card>
    </n-gi>
    <n-gi span="6 m:3">
      <n-card>
        <n-space
          justify="space-between"
          align="center"
        >
          <n-statistic :label="t('monitor.downloads')">
            <n-number-animation
              :from="0"
              :to="12039"
              show-separator
            />
          </n-statistic>
          <n-icon
            color="#ffb549"
            size="42"
          >
            <icon-park-outline-chart-graph />
          </n-icon>
        </n-space>
        <template #footer>
          <n-space justify="space-between">
            <span>{{ t('monitor.totalDownloads') }}</span>
            <span><n-number-animation
              :from="0"
              :to="322039"
              show-separator
            /></span>
          </n-space>
        </template>
      </n-card>
    </n-gi>
    <n-gi span="6 m:3">
      <n-card>
        <n-space
          justify="space-between"
          align="center"
        >
          <n-statistic :label="t('monitor.pageViews')">
            <n-number-animation
              :from="0"
              :to="12039"
              show-separator
            />
          </n-statistic>
          <n-icon
            color="#1687a7"
            size="42"
          >
            <icon-park-outline-average />
          </n-icon>
        </n-space>
        <template #footer>
          <n-space justify="space-between">
            <span>{{ t('monitor.totalPageViews') }}</span>
            <span><n-number-animation
              :from="0"
              :to="322039"
              show-separator
            /></span>
          </n-space>
        </template>
      </n-card>
    </n-gi>
    <n-gi span="6 m:3">
      <n-card>
        <n-space
          justify="space-between"
          align="center"
        >
          <n-statistic :label="t('monitor.registrations')">
            <n-number-animation
              :from="0"
              :to="12039"
              show-separator
            />
          </n-statistic>
          <n-icon
            color="#42218E"
            size="42"
          >
            <icon-park-outline-chart-pie />
          </n-icon>
        </n-space>
        <template #footer>
          <n-space justify="space-between">
            <span>{{ t('monitor.totalRegistrations') }}</span>
            <span><n-number-animation
              :from="0"
              :to="322039"
              show-separator
            /></span>
          </n-space>
        </template>
      </n-card>
    </n-gi>
    <!-- 图表区域 - 全宽显示 -->
    <n-gi :span="12">
      <n-card content-style="padding: 0;">
        <n-tabs
          type="line"
          size="large"
          :tabs-padding="20"
          pane-style="padding: 20px;"
        >
          <n-tab-pane :name="t('monitor.trafficTrend')">
            <Chart />
          </n-tab-pane>
          <n-tab-pane :name="t('monitor.visitTrend')">
            <Chart2 />
          </n-tab-pane>
        </n-tabs>
      </n-card>
    </n-gi>

    <!-- 访问来源 - 移动端全宽，桌面端1/3宽 -->
    <n-gi span="12 m:4">
      <n-card
        :title="t('monitor.visitSource')"
        :segmented="{
          content: true,
        }"
      >
        <Chart3 />
      </n-card>
    </n-gi>

    <!-- 成交记录 - 移动端全宽，桌面端2/3宽 -->
    <n-gi span="12 m:8">
      <n-card
        :title="t('monitor.transactionRecords')"
        :segmented="{
          content: true,
        }"
      >
        <template #header-extra>
          <n-button
            type="primary"
            quaternary
          >
            {{ t('monitor.more') }}
          </n-button>
        </template>
        <n-table
          :bordered="false"
          :single-line="false"
          :scroll-x="appStore.isMobile ? 600 : undefined"
        >
          <thead>
            <tr>
              <th>{{ t('monitor.transactionName') }}</th>
              <th>{{ t('monitor.startTime') }}</th>
              <th>{{ t('monitor.endTime') }}</th>
              <th>{{ t('monitor.progress') }}</th>
              <th>{{ t('monitor.status') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in tableData"
              :key="item.id"
            >
              <td>{{ item.name }}</td>
              <td>{{ item.start }}</td>
              <td>{{ item.end }}</td>
              <td>{{ item.prograss }}%</td>
              <td>
                <n-tag
                  :bordered="false"
                  type="info"
                >
                  {{ item.status }}
                </n-tag>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-card>
    </n-gi>
  </n-grid>
</template>
