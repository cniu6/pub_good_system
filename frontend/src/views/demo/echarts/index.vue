<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useEcharts } from '@/hooks'
import type { ECOption } from '@/hooks'
import { graphic } from 'echarts'

const { t } = useI18n()

// Pie chart
const pieOptions = ref<ECOption>({
  title: {
    text: t('demo.echarts.pieTitle'),
  },
  color: [
    '#37a2da',
    '#32c5e9',
    '#9fe6b8',
    '#ffdb5c',
    '#ff9f7f',
    '#fb7293',
    '#e7bcf3',
    '#8378ea',
  ],
  tooltip: {
    trigger: 'item',
    formatter: '{a} <br/>{b} : {c} ({d}%)',
  },
  toolbox: {
    show: true,
    feature: {
      mark: { show: true },
      dataView: { show: true, readOnly: false },
      magicType: {
        show: true,
        type: ['line'],
      },
      restore: { show: true },
      saveAsImage: { show: true },
    },
  },
  calculable: true,
  series: [
    {
      type: 'pie',
      name: t('demo.echarts.pieName'),
      radius: [40, 150],
      roseType: 'area',
      data: [
        { value: 10, name: 'rose1' },
        { value: 5, name: 'rose2' },
        { value: 15, name: 'rose3' },
        { value: 25, name: 'rose4' },
        { value: 20, name: 'rose5' },
        { value: 35, name: 'rose6' },
        { value: 30, name: 'rose7' },
        { value: 40, name: 'rose8' },
      ],
    },
  ],
})
const { update } = useEcharts('pieRef', pieOptions)

const randomValue = () => Math.round(Math.random() * 100)
function updatePieChart() {
  pieOptions.value.series = [
    {
      data: [
        { value: randomValue(), name: 'rose1' },
        { value: randomValue(), name: 'rose2' },
        { value: randomValue(), name: 'rose3' },
        { value: randomValue(), name: 'rose4' },
        { value: randomValue(), name: 'rose5' },
        { value: randomValue(), name: 'rose6' },
        { value: randomValue(), name: 'rose7' },
        { value: randomValue(), name: 'rose8' },
      ],
    },
  ]

  update(pieOptions.value)
}

// Line chart
const lineOptions = ref<ECOption>({
  title: {
    text: t('demo.echarts.lineTitle'),
  },
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'cross',
      label: {
        backgroundColor: '#6a7985',
      },
    },
  },
  legend: {
    data: ['Email', 'Union Ads', 'Video Ads', 'Direct', 'Search Engine'],
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true,
  },
  toolbox: {
    feature: {
      saveAsImage: {},
    },
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
  },
  yAxis: {
    type: 'value',
  },
  series: [
    {
      color: '#37a2da',
      name: 'Email',
      type: 'line',
      smooth: true,
      stack: 'Total',
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            {
              offset: 0.25,
              color: '#37a2da',
            },
            {
              offset: 1,
              color: '#fff',
            },
          ],
        },
      },
      emphasis: {
        focus: 'series',
      },
      data: [120, 132, 101, 134, 90, 230, 210],
    },
    {
      color: '#9fe6b8',
      name: 'Union Ads',
      type: 'line',
      smooth: true,
      stack: 'Total',
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            {
              offset: 0.25,
              color: '#9fe6b8',
            },
            {
              offset: 1,
              color: '#fff',
            },
          ],
        },
      },
      emphasis: {
        focus: 'series',
      },
      data: [220, 182, 191, 234, 290, 330, 310],
    },
    {
      color: '#fedb5c',
      name: 'Video Ads',
      type: 'line',
      smooth: true,
      stack: 'Total',
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            {
              offset: 0.25,
              color: '#fedb5c',
            },
            {
              offset: 1,
              color: '#fff',
            },
          ],
        },
      },
      emphasis: {
        focus: 'series',
      },
      data: [150, 232, 201, 154, 190, 330, 410],
    },
    {
      color: '#fb7293',
      name: 'Direct',
      type: 'line',
      smooth: true,
      stack: 'Total',
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            {
              offset: 0.25,
              color: '#fb7293',
            },
            {
              offset: 1,
              color: '#fff',
            },
          ],
        },
      },
      emphasis: {
        focus: 'series',
      },
      data: [320, 332, 301, 334, 390, 330, 320],
    },
    {
      color: '#e7bcf3',
      name: 'Search Engine',
      type: 'line',
      smooth: true,
      stack: 'Total',
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            {
              offset: 0.25,
              color: '#e7bcf3',
            },
            {
              offset: 1,
              color: '#fff',
            },
          ],
        },
      },
      emphasis: {
        focus: 'series',
      },
      data: [820, 932, 901, 934, 1290, 1330, 1320],
    },
  ],
})
useEcharts('lineRef', lineOptions)

// Bar chart
const barOptions = ref<ECOption>({
  title: {
    text: t('demo.echarts.barTitle'),
  },
  tooltip: {},
  grid: {
    top: '8%',
    left: '1%',
    right: '1%',
    bottom: '8%',
    containLabel: true,
  },
  legend: {
    itemGap: 50,
    data: [t('demo.echarts.totalReg'), t('demo.echarts.latestReg')],
  },
  xAxis: [
    {
      type: 'category',
      boundaryGap: true,
      axisLine: {
        // Axis line settings. Mathematical x-axis
        show: true,
        lineStyle: {
          color: '#f9f9f9',
        },
      },
      axisLabel: {
        // Axis tick label settings
        color: '#d1e6eb',
        margin: 15,
      },
      axisTick: {
        show: false,
      },
      data: [t('demo.echarts.jan'), t('demo.echarts.feb'), t('demo.echarts.mar'), t('demo.echarts.apr'), t('demo.echarts.may'), t('demo.echarts.jun'), t('demo.echarts.jul')],
    },
  ],
  yAxis: [
    {
      type: 'value',
      min: 0,
      // max: 140,
      splitNumber: 7,
      splitLine: {
        show: true,
        lineStyle: {
          color: '#0a3256',
        },
      },
      axisLine: {
        show: false,
      },
      axisLabel: {
        margin: 20,
        color: '#d1e6eb',
      },
      axisTick: {
        show: false,
      },
    },
  ],
  series: [
    {
      name: t('demo.echarts.totalReg'),
      type: 'line',
      showAllSymbol: true,
      symbol: 'emptyCircle',
      symbolSize: 6,
      itemStyle: {
        color: '#28ffb3', // line color
        borderColor: '#f0f',
      },
      label: {
        show: true,
        position: 'top',
        color: '#f9f9f9',
        borderColor: '#fff',
      },
      tooltip: {
        show: false,
      },
      areaStyle: {
        // Area fill style
        // Linear gradient, first 4 params are x0,y0,x2,y2 (range 0~1); relative to bounding box. If last param is 'true', they're absolute pixel positions.
        color: new graphic.LinearGradient(
          0,
          0,
          0,
          1,
          [
            {
              offset: 0,
              color: 'rgba(0,154,120,1)',
            },
            {
              offset: 1,
              color: 'rgba(0,0,0, 0)',
            },
          ],
          false,
        ),
        shadowColor: 'rgba(53,142,215, 0.9)', // shadow color
        shadowBlur: 20, // shadowBlur sets the blur size of shape shadow. Use with shadowColor, shadowOffsetX/Y.
      },
      data: [393, 438, 485, 631, 689, 824, 987],
    },
    {
      name: t('demo.echarts.latestReg'),
      type: 'bar',
      barWidth: 20,
      tooltip: {
        show: false,
      },
      label: {
        show: true,
        position: 'top',
        color: '#fff',
      },
      itemStyle: {
        color(params) {
          const colorList = [
            '#0ec1ff',
            '#10cdff',
            '#12daff',
            '#15ebff',
            '#17f8ff',
            '#1cfffb',
            '#1dfff1',
          ]
          return colorList[params.dataIndex]
        },
      },
      data: [200, 382, 102, 267, 186, 315, 316],
    },
  ],
})
useEcharts('barRef', barOptions)

// Radar chart
const radarOptions = ref<ECOption>({
  title: {
    text: 'Multiple Radar',
  },
  tooltip: {},
  legend: {
    left: 'center',
  },
  radar: [
    {
      indicator: [
        { name: t('demo.echarts.station1'), max: 100 },
        { name: t('demo.echarts.station2'), max: 100 },
        { name: t('demo.echarts.station3'), max: 100 },
        { name: t('demo.echarts.station4'), max: 100 },
        { name: t('demo.echarts.station5'), max: 100 },
        { name: t('demo.echarts.station6'), max: 100 },
      ],
    },
  ],
  series: [
    {
      type: 'radar',
      data: [
        {
          name: t('demo.echarts.entry'),
          value: [43, 90, 80, 53, 78, 89, 77, 50],
          areaStyle: {
            color: {
              type: 'radial',
              x: 0.5,
              y: 0.5,
              r: 0.5,
              colorStops: [
                {
                  offset: 0,
                  color: 'rgba(46,203,255, 0.14)', // color at 0%
                },
                {
                  offset: 0.15,
                  color: 'rgba(46,203,255, 0.14)', // color at 100%
                },
                {
                  offset: 0.75,
                  color: '#057FB3', // color at 100%
                },
                {
                  offset: 1,
                  color: '#078DC6', // color at 100%
                },
              ],
              global: false, // 缺省为 false
            },
          },
        },
        {
          name: t('demo.echarts.exit'),
          value: [50, 44, 56, 69, 43, 77, 90, 20],
          areaStyle: {
            color: {
              type: 'radial',
              x: 0.5,
              y: 0.5,
              r: 0.5,
              colorStops: [
                {
                  offset: 0,
                  color: 'rgba(255, 127,0, 0.14)', // color at 0%
                },
                {
                  offset: 0.15,
                  color: 'rgba(255, 127,0, 0.14)', // color at 100%
                },
                {
                  offset: 0.75,
                  color: 'rgba(2255, 127,0, 0.4)', // color at 100%
                },
                {
                  offset: 1,
                  color: 'rgba(255, 127,0, 0.5)', // color at 100%
                },
              ],
              global: false, // 缺省为 false
            },
          },
        },
      ],
    },
  ],
})
useEcharts('radarRef', radarOptions)
</script>

<template>
  <n-space
    :vertical="true"
    :size="16"
  >
    <n-card>
      <n-button @click="updatePieChart">
        {{ t('demo.echarts.updateChart') }}
      </n-button>
      <div
        ref="pieRef"
        class="h-400px"
      />
    </n-card>
    <n-card>
      <div
        ref="lineRef"
        class="h-400px"
      />
    </n-card>
    <n-card>
      <div
        ref="barRef"
        class="h-400px"
      />
    </n-card>
    <n-card>
      <div
        ref="radarRef"
        class="h-400px"
      />
    </n-card>
  </n-space>
</template>

<style scoped></style>
