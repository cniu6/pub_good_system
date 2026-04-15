<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineOptions({
  name: 'PublicAccess',
})

// 统计数据
const animatedStats = ref({
  users: 0,
  projects: 0,
  components: 0,
  sales: 0,
  orders: 0,
})

// 主题数据
const themes = [
  { name: t('landing.themeGreen'), primary: '#56CB46', secondary: '#4CAF50' },
  { name: t('landing.themeBlue'), primary: '#3b82f6', secondary: '#1d4ed8' },
  { name: t('landing.themePurple'), primary: '#8b5cf6', secondary: '#7c3aed' },
  { name: t('landing.themeOrange'), primary: '#f59e0b', secondary: '#d97706' },
  { name: t('landing.themePink'), primary: '#ec4899', secondary: '#db2777' },
]

const currentTheme = ref(themes[0])

// 动画统计数据
function animateNumber(target: number, key: keyof typeof animatedStats.value, duration = 2000) {
  const start = animatedStats.value[key]
  const startTime = performance.now()

  const animate = (currentTime: number) => {
    const elapsed = currentTime - startTime
    const progress = Math.min(elapsed / duration, 1)

    // easeOutQuart easing function
    const easeProgress = 1 - (1 - progress) ** 4
    const currentValue = Math.floor(start + (target - start) * easeProgress)

    animatedStats.value[key] = currentValue

    if (progress < 1) {
      requestAnimationFrame(animate)
    }
  }

  requestAnimationFrame(animate)
}

// 刷新数据
function refreshData() {
  animateNumber(Math.floor(Math.random() * 500) + 800, 'sales')
  animateNumber(Math.floor(Math.random() * 100) + 150, 'orders')
}

// 页面加载时的动画
onMounted(() => {
  setTimeout(() => {
    animateNumber(1200, 'users')
  }, 300)

  setTimeout(() => {
    animateNumber(850, 'projects')
  }, 600)

  setTimeout(() => {
    animateNumber(450, 'components')
  }, 900)

  setTimeout(() => {
    animateNumber(324, 'sales')
  }, 1200)

  setTimeout(() => {
    animateNumber(89, 'orders')
  }, 1500)
})
</script>

<template>
  <div class="min-h-screen bg-white">
    <!-- Navigation -->
    <nav class="relative z-50 bg-white shadow-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center py-4">
          <div class="flex items-center space-x-4">
            <div class="w-10 h-10 flex items-center justify-center">
              <img src="/favicon.svg" alt="Nova Admin Logo" class="w-10 h-10">
            </div>
            <span class="text-2xl font-bold text-gray-900">{{ t('landing.brandName') }}</span>
          </div>
          <div class="hidden md:flex items-center space-x-8">
            <a href="#features" class="text-gray-600 hover:text-gray-900 transition-colors">{{ t('landing.navFeatures') }}</a>
            <a href="#demo" class="text-gray-600 hover:text-gray-900 transition-colors">{{ t('landing.navDemo') }}</a>
            <router-link
              to="/login"
              class="text-white px-6 py-2 rounded-full hover:shadow-lg transition-all transform hover:scale-105"
              style="background: linear-gradient(to right, #56CB46, #4CAF50);"
            >
              {{ t('landing.navLogin') }}
            </router-link>
          </div>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <section class="relative overflow-hidden bg-gradient-to-br from-green-50 via-white to-emerald-50">
      <div class="absolute inset-0 bg-grid-pattern opacity-5" />
      <div class="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
        <div class="text-center">
          <h1 class="text-5xl md:text-7xl font-extrabold text-gray-900 mb-8 leading-tight">
            {{ t('landing.heroTitle1') }}
            <span class="bg-clip-text text-transparent" style="background: linear-gradient(to right, #56CB46, #4CAF50); -webkit-background-clip: text; background-clip: text;">
              {{ t('landing.heroTitle2') }}
            </span>
          </h1>
          <p class="text-xl md:text-2xl text-gray-600 mb-12 max-w-3xl mx-auto leading-relaxed">
            {{ t('landing.heroDesc') }}
          </p>
          <div class="flex flex-col sm:flex-row justify-center items-center space-y-4 sm:space-y-0 sm:space-x-6">
            <button class="text-white px-8 py-4 rounded-full text-lg font-semibold hover:shadow-xl transition-all transform hover:scale-105" style="background: linear-gradient(to right, #56CB46, #4CAF50);">
              {{ t('landing.ctaTryFree') }}
            </button>
            <button class="border-2 border-gray-300 text-gray-700 px-8 py-4 rounded-full text-lg font-semibold hover:border-gray-400 transition-colors">
              {{ t('landing.ctaViewDemo') }}
            </button>
          </div>
        </div>

        <!-- Stats -->
        <div class="mt-20 grid grid-cols-2 md:grid-cols-4 gap-8">
          <div class="text-center">
            <div class="text-4xl font-bold text-gray-900 mb-2">
              {{ animatedStats.users }}+
            </div>
            <div class="text-gray-600">
              {{ t('landing.statActiveUsers') }}
            </div>
          </div>
          <div class="text-center">
            <div class="text-4xl font-bold text-gray-900 mb-2">
              {{ animatedStats.projects }}+
            </div>
            <div class="text-gray-600">
              {{ t('landing.statDeployments') }}
            </div>
          </div>
          <div class="text-center">
            <div class="text-4xl font-bold text-gray-900 mb-2">
              {{ animatedStats.components }}+
            </div>
            <div class="text-gray-600">
              {{ t('landing.statComponents') }}
            </div>
          </div>
          <div class="text-center">
            <div class="text-4xl font-bold text-gray-900 mb-2">
              99.9%
            </div>
            <div class="text-gray-600">
              {{ t('landing.statUptime') }}
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features -->
    <section id="features" class="py-24 bg-gray-50">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="text-center mb-16">
          <h2 class="text-4xl font-bold text-gray-900 mb-4">
            {{ t('landing.featuresTitle') }}
          </h2>
          <p class="text-xl text-gray-600 max-w-3xl mx-auto">
            {{ t('landing.featuresSubtitle') }}
          </p>
        </div>

        <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          <div class="bg-white rounded-2xl p-8 shadow-sm hover:shadow-lg transition-shadow">
            <div class="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center mb-6">
              <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-gray-900 mb-3">
              {{ t('landing.feature1Title') }}
            </h3>
            <p class="text-gray-600">
              {{ t('landing.feature1Desc') }}
            </p>
          </div>

          <div class="bg-white rounded-2xl p-8 shadow-sm hover:shadow-lg transition-shadow">
            <div class="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center mb-6">
              <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-gray-900 mb-3">
              {{ t('landing.feature2Title') }}
            </h3>
            <p class="text-gray-600">
              {{ t('landing.feature2Desc') }}
            </p>
          </div>

          <div class="bg-white rounded-2xl p-8 shadow-sm hover:shadow-lg transition-shadow">
            <div class="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center mb-6">
              <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-gray-900 mb-3">
              {{ t('landing.feature3Title') }}
            </h3>
            <p class="text-gray-600">
              {{ t('landing.feature3Desc') }}
            </p>
          </div>

          <div class="bg-white rounded-2xl p-8 shadow-sm hover:shadow-lg transition-shadow">
            <div class="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center mb-6">
              <svg class="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-gray-900 mb-3">
              {{ t('landing.feature4Title') }}
            </h3>
            <p class="text-gray-600">
              {{ t('landing.feature4Desc') }}
            </p>
          </div>

          <div class="bg-white rounded-2xl p-8 shadow-sm hover:shadow-lg transition-shadow">
            <div class="w-12 h-12 bg-indigo-100 rounded-xl flex items-center justify-center mb-6">
              <svg class="w-6 h-6 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-gray-900 mb-3">
              {{ t('landing.feature5Title') }}
            </h3>
            <p class="text-gray-600">
              {{ t('landing.feature5Desc') }}
            </p>
          </div>

          <div class="bg-white rounded-2xl p-8 shadow-sm hover:shadow-lg transition-shadow">
            <div class="w-12 h-12 bg-pink-100 rounded-xl flex items-center justify-center mb-6">
              <svg class="w-6 h-6 text-pink-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
              </svg>
            </div>
            <h3 class="text-xl font-semibold text-gray-900 mb-3">
              {{ t('landing.feature6Title') }}
            </h3>
            <p class="text-gray-600">
              {{ t('landing.feature6Desc') }}
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Demo Section -->
    <section id="demo" class="py-24 bg-white">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="text-center mb-16">
          <h2 class="text-4xl font-bold text-gray-900 mb-4">
            {{ t('landing.demoTitle') }}
          </h2>
          <p class="text-xl text-gray-600">
            {{ t('landing.demoSubtitle') }}
          </p>
        </div>

        <div class="grid md:grid-cols-2 gap-12 items-center">
          <!-- Demo Area -->
          <div class="space-y-8">
            <div class="bg-gray-50 rounded-2xl p-8">
              <h3 class="text-xl font-semibold text-gray-900 mb-4">
                {{ t('landing.demoPanelTitle') }}
              </h3>
              <div class="grid grid-cols-2 gap-4 mb-6">
                <div class="bg-white rounded-lg p-4 text-center">
                  <div class="text-2xl font-bold mb-1" style="color: #56CB46;">
                    {{ animatedStats.sales }}
                  </div>
                  <div class="text-sm text-gray-600">
                    {{ t('landing.demoSales') }}
                  </div>
                </div>
                <div class="bg-white rounded-lg p-4 text-center">
                  <div class="text-2xl font-bold text-green-600 mb-1">
                    {{ animatedStats.orders }}
                  </div>
                  <div class="text-sm text-gray-600">
                    {{ t('landing.demoOrders') }}
                  </div>
                </div>
              </div>
              <button
                class="w-full text-white py-2 rounded-lg transition-colors"
                style="background-color: #56CB46;"
                onmouseover="this.style.backgroundColor='#4CAF50'"
                onmouseout="this.style.backgroundColor='#56CB46'"
                @click="refreshData"
              >
                {{ t('landing.demoRefresh') }}
              </button>
            </div>

            <div class="bg-gray-50 rounded-2xl p-8">
              <h3 class="text-xl font-semibold text-gray-900 mb-4">
                {{ t('landing.themeTitle') }}
              </h3>
              <div class="flex space-x-3 mb-4">
                <button
                  v-for="theme in themes"
                  :key="theme.name"
                  class="w-12 h-12 rounded-full border-2 transition-all transform hover:scale-110"
                  :class="currentTheme.name === theme.name ? 'border-gray-400 scale-110' : 'border-gray-200'"
                  :style="{ background: `linear-gradient(135deg, ${theme.primary}, ${theme.secondary})` }"
                  @click="currentTheme = theme"
                />
              </div>
              <div
                class="h-24 rounded-lg transition-all duration-500 flex items-center justify-center text-white font-semibold"
                :style="{ background: `linear-gradient(135deg, ${currentTheme.primary}, ${currentTheme.secondary})` }"
              >
                {{ currentTheme.name }} {{ t('landing.themeSuffix') }}
              </div>
            </div>
          </div>

          <!-- Feature List -->
          <div class="space-y-6">
            <div class="flex items-start space-x-4">
              <div class="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center flex-shrink-0 mt-1">
                <svg class="w-4 h-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                </svg>
              </div>
              <div>
                <h4 class="text-lg font-semibold text-gray-900">
                  {{ t('landing.checkI18n') }}
                </h4>
                <p class="text-gray-600">
                  {{ t('landing.checkI18nDesc') }}
                </p>
              </div>
            </div>

            <div class="flex items-start space-x-4">
              <div class="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center flex-shrink-0 mt-1">
                <svg class="w-4 h-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                </svg>
              </div>
              <div>
                <h4 class="text-lg font-semibold text-gray-900">
                  {{ t('landing.checkDark') }}
                </h4>
                <p class="text-gray-600">
                  {{ t('landing.checkDarkDesc') }}
                </p>
              </div>
            </div>

            <div class="flex items-start space-x-4">
              <div class="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center flex-shrink-0 mt-1">
                <svg class="w-4 h-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                </svg>
              </div>
              <div>
                <h4 class="text-lg font-semibold text-gray-900">
                  {{ t('landing.checkCharts') }}
                </h4>
                <p class="text-gray-600">
                  {{ t('landing.checkChartsDesc') }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="py-24" style="background: linear-gradient(to right, #56CB46, #4CAF50);">
      <div class="max-w-4xl mx-auto text-center px-4 sm:px-6 lg:px-8">
        <h2 class="text-4xl font-bold text-white mb-6">
          {{ t('landing.ctaTitle') }}
        </h2>
        <p class="text-xl text-green-100 mb-10 max-w-2xl mx-auto">
          {{ t('landing.ctaSubtitle') }}
        </p>
        <div class="flex flex-col sm:flex-row justify-center items-center space-y-4 sm:space-y-0 sm:space-x-6">
          <router-link
            to="/login"
            class="bg-white px-8 py-4 rounded-full text-lg font-semibold hover:shadow-lg transition-all transform hover:scale-105"
            style="color: #56CB46;"
          >
            {{ t('landing.ctaStartNow') }}
          </router-link>
          <a
            href="https://github.com"
            target="_blank"
            class="border-2 border-white text-white px-8 py-4 rounded-full text-lg font-semibold hover:bg-white transition-all"
            style="transition: all 0.3s ease;"
            onmouseover="this.style.color='#56CB46'"
            onmouseout="this.style.color='white'"
          >
            {{ t('landing.ctaViewSource') }}
          </a>
        </div>
      </div>
    </section>

    <!-- Footer -->
    <footer class="bg-gray-900 text-gray-300 py-12">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="grid md:grid-cols-4 gap-8">
          <div>
            <div class="flex items-center space-x-2 mb-4">
              <div class="w-10 h-10 flex items-center justify-center">
                <img src="/favicon.svg" :alt="t('landing.brandName') + ' Logo'" class="w-8 h-8">
              </div>
              <span class="text-xl font-bold text-white">{{ t('landing.brandName') }}</span>
            </div>
            <p class="text-gray-400">
              {{ t('landing.footerDesc') }}
            </p>
          </div>

          <div>
            <h3 class="text-white font-semibold mb-4">
              {{ t('landing.footerProduct') }}
            </h3>
            <ul class="space-y-2">
              <li><a href="#" class="hover:text-white transition-colors">{{ t('landing.footerFeatures') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ t('landing.footerPricing') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ t('landing.footerChangelog') }}</a></li>
            </ul>
          </div>

          <div>
            <h3 class="text-white font-semibold mb-4">
              {{ t('landing.footerSupport') }}
            </h3>
            <ul class="space-y-2">
              <li><a href="#" class="hover:text-white transition-colors">{{ t('landing.footerDocs') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ t('landing.footerApi') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ t('landing.footerCommunity') }}</a></li>
            </ul>
          </div>

          <div>
            <h3 class="text-white font-semibold mb-4">
              {{ t('landing.footerCompany') }}
            </h3>
            <ul class="space-y-2">
              <li><a href="#" class="hover:text-white transition-colors">{{ t('landing.footerAbout') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ t('landing.footerContact') }}</a></li>
              <li><a href="#" class="hover:text-white transition-colors">{{ t('landing.footerPrivacy') }}</a></li>
            </ul>
          </div>
        </div>

        <div class="border-t border-gray-800 mt-8 pt-8 text-center">
          <p class="text-gray-400">
            © 2024 {{ t('landing.brandName') }}. {{ t('landing.footerRights') }}
          </p>
        </div>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.bg-grid-pattern {
  background-image: radial-gradient(circle, #e2e8f0 1px, transparent 1px);
  background-size: 20px 20px;
}

/* 平滑滚动 */
html {
  scroll-behavior: smooth;
}

/* 自定义渐变文字 */
.bg-clip-text {
  -webkit-background-clip: text;
  background-clip: text;
}
</style>
