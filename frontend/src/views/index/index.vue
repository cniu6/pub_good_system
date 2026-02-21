<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/store'
import { local } from '@/utils'
import { useI18n } from 'vue-i18n'

defineOptions({ name: 'IndexPage' })

const router = useRouter()
const appStore = useAppStore()
const { t } = useI18n()

// 原生深色模式
const is_dark = computed(() => appStore.colorMode === 'dark')

// 清除 Naive UI 的内联背景色
let _original_bg = ''
function clear_naive_bg() {
  const el = document.querySelector('.n-config-provider') as HTMLElement
  if (el) {
    _original_bg = el.style.backgroundColor
    el.style.backgroundColor = 'transparent'
  }
}
function restore_naive_bg() {
  const el = document.querySelector('.n-config-provider') as HTMLElement
  if (el && _original_bg) {
    el.style.backgroundColor = _original_bg
  }
}

// 状态判断
const is_logged_in = computed(() => Boolean(local.get('accessToken')))

// 统计动画
const stats = ref({ users: 0, plugins: 0, uptime: 0, api: 0 })
function animate_all_stats() {
  const start = performance.now()
  const duration = 2500
  const tick = (now: number) => {
    const p = Math.min((now - start) / duration, 1)
    const ease = 1 - Math.pow(1 - p, 4)
    stats.value.users = Math.floor(ease * 1256)
    stats.value.plugins = Math.floor(ease * 24)
    stats.value.uptime = Math.floor(ease * 99)
    stats.value.api = Math.floor(ease * 3500)
    if (p < 1) requestAnimationFrame(tick)
  }
  requestAnimationFrame(tick)
}

// 导航
function scroll_to(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' })
}

function go_user() {
  const target = import.meta.env.VITE_HOME_PATH || '/dashboard/workbench'
  router.push(is_logged_in.value ? target : { path: '/login', query: { redirect: target } })
}

function go_admin() {
  const target = import.meta.env.VITE_ADMIN_BASE_PATH || '/system-mgr'
  router.push(is_logged_in.value ? target : { path: '/login', query: { redirect: target } })
}

// 鼠标位置跟踪 (用于光晕动画 - 使用 CSS 变量避免 reactive 开销)
const glow_style = ref({ '--glow-x': '0px', '--glow-y': '0px' })
let glow_target_x = 0
let glow_target_y = 0
let glow_current_x = 0
let glow_current_y = 0
let glow_raf_id: number | null = null

function handle_mouse(e: MouseEvent) {
  glow_target_x = e.clientX
  glow_target_y = e.clientY
}

// 平滑跟随动画 - 使用 rAF 实现丝滑效果
function animate_glow() {
  // 线性插值 (lerp) 实现平滑跟随
  const ease = 0.15
  glow_current_x += (glow_target_x - glow_current_x) * ease
  glow_current_y += (glow_target_y - glow_current_y) * ease
  
  glow_style.value = {
    '--glow-x': `${glow_current_x}px`,
    '--glow-y': `${glow_current_y}px`
  }
  glow_raf_id = requestAnimationFrame(animate_glow)
}

const is_scrolled = ref(false)
function handle_scroll() {
  is_scrolled.value = window.scrollY > 30
}

onMounted(() => {
  window.addEventListener('mousemove', handle_mouse, { passive: true })
  window.addEventListener('scroll', handle_scroll, { passive: true })
  clear_naive_bg()
  animate_all_stats()
  animate_glow() // 启动光晕平滑动画
  watch(() => appStore.colorMode, () => requestAnimationFrame(clear_naive_bg))
})

onBeforeUnmount(() => {
  window.removeEventListener('mousemove', handle_mouse)
  window.removeEventListener('scroll', handle_scroll)
  restore_naive_bg()
  // 停止光晕动画
  if (glow_raf_id) {
    cancelAnimationFrame(glow_raf_id)
    glow_raf_id = null
  }
})
</script>

<template>
  <div class="index-page" :class="{ 'index-dark': is_dark }">
    <!-- 背景层 -->
    <div class="bg-layer" :style="glow_style">
      <div class="mesh-gradient"></div>
      <div class="dot-grid"></div>
      <div class="glow-point"></div>
    </div>

    <!-- 导航栏 -->
    <nav class="nav-glass" :class="{ scrolled: is_scrolled }">
      <div class="container nav-content">
        <div class="logo-area" @click="router.push('/')">
          <div class="logo-box">F</div>
          <span class="logo-text">F.st</span>
        </div>
        <div class="nav-links">
          <a @click="scroll_to('features')">{{ t('home.nav.features') }}</a>
          <a @click="scroll_to('tech')">{{ t('home.nav.tech') }}</a>
          <a @click="scroll_to('about')">{{ t('home.nav.about') }}</a>
        </div>
        <div class="nav-btns">
          <DarkModeSwitch />
          <LangsSwitch />
          <div class="custom-btn-group">
            <button class="custom-btn btn-primary" @click="go_user">
              <svg class="btn-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
              {{ t('home.nav.userConsole') }}
            </button>
            <button class="custom-btn btn-warning" @click="go_admin">
              <svg class="btn-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
              {{ t('home.nav.adminConsole') }}
            </button>
          </div>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <section class="hero-section container">
      <div class="hero-left">
        <div class="hero-badge">{{ t('home.hero.badge') }}</div>
        <h1 class="hero-title">
          {{ t('home.hero.title') }}<br />
          <span class="gradient-text">Run F.st</span>
        </h1>
        <p class="hero-subtitle">{{ t('home.hero.subtitle') }}</p>
        <div class="hero-actions">
          <button class="hero-btn btn-primary-lg" @click="go_user">
            <svg class="btn-icon-lg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
            {{ t('home.hero.enterUser') }}
          </button>
          <button class="hero-btn btn-warning-lg" @click="go_admin">
            <svg class="btn-icon-lg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
            {{ t('home.hero.enterAdmin') }}
          </button>
          <button class="hero-btn btn-ghost-lg" @click="scroll_to('features')">
            {{ t('home.hero.learnMore') }}
          </button>
        </div>
        <div class="hero-stats">
          <div class="stat-box">
             <span class="stat-val">{{ stats.users }}+</span>
             <span class="stat-key">{{ t('home.hero.users') }}</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-box">
             <span class="stat-val">{{ stats.plugins }}</span>
             <span class="stat-key">{{ t('home.hero.plugins') }}</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-box">
             <span class="stat-val">{{ stats.uptime }}%</span>
             <span class="stat-key">{{ t('home.hero.uptime') }}</span>
          </div>
        </div>
      </div>
      
      <div class="hero-right">
        <!-- 模拟控制台窗口 -->
        <div class="mockup-window">
          <div class="mockup-header">
            <div class="dots"><span class="r"></span><span class="y"></span><span class="g"></span></div>
            <div class="mockup-tab">dashboard.go</div>
          </div>
          <div class="mockup-body">
            <pre class="code-block"><code><span class="k">package</span> main

<span class="k">import</span> <span class="s">"github.com/fst/core"</span>

<span class="k">func</span> <span class="f">Init</span>() {
  app := core.<span class="f">New</span>()
  <span class="c">// Hyper-modular plugin architecture</span>
  app.<span class="f">Use</span>(core.<span class="f">Auth</span>())
  app.<span class="f">Bootstrap</span>()
}</code></pre>
            <div class="floating-card user-card">
              <div class="avatar blue"></div>
              <div class="lines"><span></span><span></span></div>
            </div>
            <div class="floating-card chart-card">
              <div class="chart-bars"><div class="b" style="height:40%"></div><div class="b" style="height:70%"></div><div class="b" style="height:50%"></div><div class="b" style="height:90%"></div></div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Bento Grid Features -->
    <section id="features" class="features-section container">
      <div class="section-title-box">
        <h2 class="section-title">{{ t('home.features.title') }}</h2>
        <p class="section-desc">{{ t('home.features.subtitle') }}</p>
      </div>

      <div class="bento-grid">
        <!-- Main: Plugin -->
        <div class="bento-item main-card">
          <div class="card-icon blue"><svg class="icon-32" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"/></svg></div>
          <h3>{{ t('home.features.plugin') }}</h3>
          <p>{{ t('home.features.pluginDesc') }}</p>
          <div class="card-visual plug-visual">
            <div class="plug"></div>
            <div class="slots"><span></span><span></span><span></span></div>
          </div>
        </div>

        <!-- Performance -->
        <div class="bento-item">
          <div class="card-icon green"><svg class="icon-32" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/></svg></div>
          <h3>{{ t('home.features.perf') }}</h3>
          <p>{{ t('home.features.perfDesc') }}</p>
        </div>

        <!-- Security -->
        <div class="bento-item">
          <div class="card-icon orange"><svg class="icon-32" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/></svg></div>
          <h3>{{ t('home.features.security') }}</h3>
          <p>{{ t('home.features.securityDesc') }}</p>
        </div>

        <!-- MVC -->
        <div class="bento-item horizontal">
          <div class="flex-col">
            <div class="card-icon purple"><svg class="icon-32" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1-0 01-1-1v-6z"/></svg></div>
            <h3>{{ t('home.features.mvc') }}</h3>
            <p>{{ t('home.features.mvcDesc') }}</p>
          </div>
        </div>

        <!-- Deploy -->
        <div class="bento-item">
          <div class="card-icon cyan"><svg class="icon-32" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg></div>
          <h3>{{ t('home.features.deploy') }}</h3>
          <p>{{ t('home.features.deployDesc') }}</p>
        </div>
      </div>
    </section>

    <!-- Tech Stack Section -->
    <section id="tech" class="tech-section container">
      <div class="section-title-box">
        <h2 class="section-title">{{ t('home.tech.title') }}</h2>
        <p class="section-desc">{{ t('home.tech.subtitle') }}</p>
      </div>
      <div class="tech-box">
        <div class="tech-tag" v-for="t in ['Go 1.24', 'Vue 3.5', 'Vite 7', 'Pinia', 'TypeScript', 'MySQL 8', 'UnoCSS', 'Naive UI']" :key="t">
          {{ t }}
        </div>
      </div>
    </section>

    <!-- Final CTA -->
    <section id="about" class="cta-wrap container">
      <div class="cta-inner">
        <h2 class="cta-h">{{ t('home.cta.title') }}</h2>
        <p class="cta-p">{{ t('home.cta.desc') }}</p>
        <div class="cta-btn-group">
          <button class="cta-btn btn-white" @click="go_user">
            {{ t('home.cta.startUser') }}
          </button>
          <button class="cta-btn btn-outline-white" @click="go_admin">
            {{ t('home.cta.startAdmin') }}
          </button>
        </div>
      </div>
    </section>

    <!-- Footer -->
    <footer class="footer">
      <div class="container footer-grid">
        <div class="footer-info">
          <div class="logo-area mb-4">
            <div class="logo-box">F</div>
            <span class="logo-text">F.st</span>
          </div>
          <p class="opacity-60 text-sm">Empowering the next generation of full-stack enterprise applications.</p>
        </div>
        <div class="footer-links">
          <h4>Platform</h4>
          <a @click="scroll_to('features')">Features</a>
          <a @click="scroll_to('tech')">Ecosystem</a>
        </div>
        <div class="footer-links">
          <h4>Resources</h4>
          <a href="#">Documentation</a>
          <a href="#">GitHub</a>
        </div>
      </div>
      <div class="container footer-bottom">
        <p>© 2024 F.st. Built with Passion for Developers.</p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;700;800&display=swap');

.index-page {
  --primary: #10b981;
  --secondary: #3b82f6;
  --text: #111827;
  --text-soft: #4b5563;
  --bg: #ffffff;
  --panel: rgba(255, 255, 255, 0.7);
  --border: rgba(0, 0, 0, 0.08);
  --shadow: 0 10px 30px -5px rgba(0, 0, 0, 0.05);

  font-family: 'Plus Jakarta Sans', sans-serif;
  color: var(--text);
  background: var(--bg) !important;
  min-height: 100vh;
}

.index-page.index-dark {
  --text: #f9fafb;
  --text-soft: #9ca3af;
  --bg: #030712;
  --panel: rgba(17, 24, 39, 0.7);
  --border: rgba(255, 255, 255, 0.1);
  --shadow: 0 20px 40px -10px rgba(0, 0, 0, 0.5);
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
}

/* ===== BG Elements ===== */
.bg-layer {
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
}
.mesh-gradient {
  position: absolute;
  inset: 0;
  background: 
    radial-gradient(circle at 10% 10%, rgba(16, 185, 129, 0.1), transparent 40%),
    radial-gradient(circle at 90% 90%, rgba(59, 130, 246, 0.1), transparent 40%),
    radial-gradient(circle at 50% 50%, rgba(139, 92, 246, 0.05), transparent 40%);
}
.dot-grid {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(var(--border) 1px, transparent 1px);
  background-size: 30px 30px;
  mask-image: radial-gradient(circle at center, black, transparent 80%);
}
.glow-point {
  position: absolute;
  width: 400px;
  height: 400px;
  left: var(--glow-x, 0);
  top: var(--glow-y, 0);
  background: radial-gradient(circle, var(--primary), transparent 70%);
  filter: blur(80px);
  opacity: 0.1;
  transform: translate(-50%, -50%);
  will-change: left, top;
  pointer-events: none;
}

/* ===== Nav ===== */
.nav-glass {
  position: fixed;
  top: 24px;
  left: 0;
  right: 0;
  z-index: 1000;
  transition: all 0.3s;
}
.nav-glass.scrolled {
  top: 0;
  background: var(--panel);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border);
}
.nav-content {
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.logo-area {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  font-weight: 800;
  font-size: 24px;
}
.logo-box {
  width: 36px;
  height: 36px;
  background: var(--primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  box-shadow: 0 4px 10px rgba(16, 185, 129, 0.3);
}
.nav-links {
  display: flex;
  gap: 32px;
}
.nav-links a {
  color: var(--text-soft);
  font-weight: 600;
  cursor: pointer;
  transition: color 0.2s;
}
.nav-links a:hover { color: var(--primary); }
.nav-btns {
  display: flex;
  align-items: center;
  gap: 16px;
}

/* Custom Button Group */
.custom-btn-group {
  display: flex;
  gap: 8px;
}
.custom-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  border-radius: 10px;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}
.custom-btn .btn-icon {
  flex-shrink: 0;
}
.btn-primary {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
}
.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(16, 185, 129, 0.4);
}
.btn-primary:active {
  transform: translateY(0);
  box-shadow: 0 2px 6px rgba(16, 185, 129, 0.3);
}
.btn-warning {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.3);
}
.btn-warning:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(245, 158, 11, 0.4);
}
.btn-warning:active {
  transform: translateY(0);
  box-shadow: 0 2px 6px rgba(245, 158, 11, 0.3);
}


/* ===== Hero ===== */
.hero-section {
  padding-top: 180px;
  padding-bottom: 100px;
  display: flex;
  align-items: center;
  gap: 60px;
  position: relative;
  z-index: 1;
}
.hero-left { flex: 1; }
.hero-right { flex: 1; display: flex; justify-content: flex-end; }

.hero-badge {
  display: inline-block;
  padding: 6px 14px;
  background: rgba(16, 185, 129, 0.1);
  color: var(--primary);
  border: 1px solid rgba(16, 185, 129, 0.2);
  border-radius: 99px;
  font-weight: 700;
  font-size: 14px;
  margin-bottom: 24px;
}
.hero-title {
  font-size: 64px;
  font-weight: 800;
  line-height: 1.1;
  margin-bottom: 24px;
  letter-spacing: -2px;
}
.gradient-text {
  background: linear-gradient(135deg, var(--primary), var(--secondary));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}
.hero-subtitle {
  font-size: 18px;
  color: var(--text-soft);
  line-height: 1.6;
  margin-bottom: 40px;
}
.hero-actions {
  display: flex;
  gap: 16px;
  margin-bottom: 60px;
}

/* Hero Buttons */
.hero-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0 28px;
  height: 48px;
  font-size: 16px;
  font-weight: 700;
  border-radius: 14px;
  border: none;
  cursor: pointer;
  transition: all 0.25s ease;
  white-space: nowrap;
}
.btn-icon-lg {
  flex-shrink: 0;
}
.btn-primary-lg {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  box-shadow: 0 4px 20px rgba(16, 185, 129, 0.35);
}
.btn-primary-lg:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 28px rgba(16, 185, 129, 0.45);
}
.btn-primary-lg:active {
  transform: translateY(0);
  box-shadow: 0 4px 14px rgba(16, 185, 129, 0.35);
}
.btn-warning-lg {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  box-shadow: 0 4px 20px rgba(245, 158, 11, 0.35);
}
.btn-warning-lg:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 28px rgba(245, 158, 11, 0.45);
}
.btn-warning-lg:active {
  transform: translateY(0);
  box-shadow: 0 4px 14px rgba(245, 158, 11, 0.35);
}
.btn-ghost-lg {
  background: transparent;
  color: var(--text-soft);
  border: 2px solid var(--border);
}
.btn-ghost-lg:hover {
  background: var(--panel);
  border-color: var(--primary);
  color: var(--primary);
}


.hero-stats {
  display: flex;
  align-items: center;
  gap: 40px;
}
.stat-box { display: flex; flex-direction: column; }
.stat-val { font-size: 32px; font-weight: 800; }
.stat-key { font-size: 14px; color: var(--text-soft); }
.stat-divider { width: 1px; height: 40px; background: var(--border); }

/* Mockup */
.mockup-window {
  width: 100%;
  max-width: 540px;
  background: var(--panel);
  backdrop-filter: blur(20px);
  border: 1px solid var(--border);
  border-radius: 24px;
  box-shadow: var(--shadow);
  overflow: hidden;
  position: relative;
}
.mockup-header {
  height: 48px;
  background: var(--border);
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 16px;
}
.dots { display: flex; gap: 8px; }
.dots span { width: 12px; height: 12px; border-radius: 50%; opacity: 0.8; }
.dots .r { background: #ff5f56; }
.dots .y { background: #ffbd2e; }
.dots .g { background: #27c93f; }
.mockup-tab {
  background: var(--bg);
  padding: 6px 16px;
  border-radius: 8px 8px 0 0;
  font-size: 12px;
  font-weight: 600;
  position: relative;
  top: 10px;
}
.mockup-body {
  padding: 24px;
  position: relative;
  min-height: 240px;
}
.code-block {
  font-family: 'JetBrains Mono', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: var(--text-soft);
}
.k { color: #8b5cf6; } /* Keyword */
.s { color: #10b981; } /* String */
.f { color: #3b82f6; } /* Function */
.c { color: #94a3b8; font-style: italic; } /* Comment */

.floating-card {
  position: absolute;
  padding: 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 16px;
  box-shadow: var(--shadow);
}
.user-card { bottom: 20px; right: 20px; display: flex; align-items: center; gap: 12px; width: 160px; }
.avatar { width: 32px; height: 32px; border-radius: 50%; }
.avatar.blue { background: var(--secondary); }
.lines span { display: block; height: 6px; background: var(--border); border-radius: 3px; margin: 4px 0; }
.lines span:first-child { width: 60px; }
.lines span:last-child { width: 40px; }
.chart-card { top: 60px; right: -20px; width: 100px; padding: 16px; }
.chart-bars { display: flex; align-items: flex-end; gap: 4px; height: 60px; }
.chart-bars .b { flex: 1; background: var(--primary); border-radius: 4px; }

/* ===== Bento Grid ===== */
.features-section { padding-top: 100px; padding-bottom: 150px; }
.section-title-box { margin-bottom: 60px; text-align: center; }
.section-title { font-size: 48px; font-weight: 800; margin-bottom: 16px; }
.section-desc { font-size: 18px; color: var(--text-soft); }

.bento-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  grid-auto-rows: 240px;
  gap: 24px;
}
.bento-item {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 32px;
  padding: 40px;
  transition: all 0.3s;
  position: relative;
  overflow: hidden;
}
.bento-item:hover { border-color: var(--primary); background: var(--bg-card); }
.main-card { grid-column: span 2; grid-row: span 2; display: flex; flex-direction: column; }
.bento-item h3 { font-size: 24px; font-weight: 800; margin-bottom: 16px; }
.bento-item p { font-size: 16px; color: var(--text-soft); line-height: 1.6; }
.horizontal { grid-column: span 2; display: flex; align-items: center; }

.card-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
}
.card-icon.blue { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.card-icon.green { background: rgba(16, 185, 129, 0.1); color: #10b981; }
.card-icon.orange { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }
.card-icon.purple { background: rgba(139, 92, 246, 0.1); color: #8b5cf6; }
.card-icon.cyan { background: rgba(6, 182, 212, 0.1); color: #06b6d4; }

.plug-visual {
  margin-top: auto;
  padding-top: 40px;
  display: flex;
  justify-content: center;
  gap: 20px;
}
.plug { width: 80px; height: 30px; background: var(--primary); border-radius: 15px; }
.slots span { display: inline-block; width: 40px; height: 10px; background: var(--border); border-radius: 5px; margin: 0 4px; }

/* ===== Tech Stack ===== */
.tech-box {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 12px;
  margin-top: 40px;
}
.tech-tag {
  padding: 12px 24px;
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 16px;
  font-weight: 700;
  font-size: 16px;
  transition: all 0.2s;
}
.tech-tag:hover { border-color: var(--primary); color: var(--primary); transform: scale(1.05); }

/* ===== CTA ===== */
.cta-wrap { padding-top: 100px; padding-bottom: 150px; }
.cta-inner {
  background: black;
  color: white;
  padding: 80px;
  border-radius: 48px;
  text-align: center;
  position: relative;
  overflow: hidden;
}
.index-dark .cta-inner { background: #111827; border: 1px solid var(--border); }
.cta-h { font-size: 56px; font-weight: 800; margin-bottom: 24px; letter-spacing: -2px; }
.cta-p { font-size: 20px; opacity: 0.7; margin-bottom: 48px; max-width: 600px; margin-inline: auto; }

/* CTA Buttons */
.cta-btn-group {
  display: flex;
  justify-content: center;
  gap: 16px;
}
.cta-btn {
  padding: 0 36px;
  height: 52px;
  font-size: 17px;
  font-weight: 700;
  border-radius: 16px;
  border: none;
  cursor: pointer;
  transition: all 0.25s ease;
  white-space: nowrap;
}
.btn-white {
  background: white;
  color: #111827;
  box-shadow: 0 4px 20px rgba(255, 255, 255, 0.2);
}
.btn-white:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 30px rgba(255, 255, 255, 0.3);
}
.btn-white:active {
  transform: translateY(0);
}
.btn-outline-white {
  background: transparent;
  color: white;
  border: 2px solid rgba(255, 255, 255, 0.4);
}
.btn-outline-white:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.7);
}

/* ===== Footer ===== */
.footer { border-top: 1px solid var(--border); padding: 80px 0 40px; }
.footer-grid { display: grid; grid-template-columns: 2fr 1fr 1fr; gap: 80px; margin-bottom: 80px; }
.footer h4 { font-weight: 800; margin-bottom: 24px; font-size: 14px; text-transform: uppercase; letter-spacing: 1px; }
.footer a { display: block; margin-bottom: 12px; color: var(--text-soft); font-weight: 600; text-decoration: none; cursor: pointer; }
.footer a:hover { color: var(--primary); }
.footer-bottom { border-top: 1px solid var(--border); padding-top: 40px; color: var(--text-soft); font-size: 14px; font-weight: 600; }

.icon-20 { width: 20px; height: 20px; }
.icon-32 { width: 32px; height: 32px; }
.mb-4 { margin-bottom: 16px; }
.opacity-60 { opacity: 0.6; }
.text-sm { font-size: 14px; }
.flex-col { display: flex; flex-direction: column; }

@media (max-width: 1024px) {
  .hero-section { flex-direction: column; text-align: center; }
  .hero-right { justify-content: center; transform: scale(0.9); }
  .hero-actions { justify-content: center; }
  .hero-stats { justify-content: center; }
  .bento-grid { grid-template-columns: 1fr; grid-auto-rows: auto; }
  .main-card { grid-column: auto; grid-row: auto; }
  .horizontal { grid-column: auto; }
  .footer-grid { grid-template-columns: 1fr; gap: 40px; }
}

@media (max-width: 768px) {
  .hero-title { font-size: 48px; }
  .cta-inner { padding: 40px 24px; }
  .cta-h { font-size: 36px; }
}
</style>
