/**
 * 启动时在浏览器控制台打印「品牌 + 编译时间」双徽章样式日志。
 * 用 console 的 %c 注入 CSS，左右拼成一条 pill；先 clear 再打，保证排在控制台最前。
 */
export function printBuildBanner(): void {
  const name = import.meta.env.VITE_APP_NAME || 'F.st'
  const buildTime = typeof __BUILD_TIMESTAMP__ !== 'undefined'
    ? __BUILD_TIMESTAMP__
    : 'dev'

  // inline-block 才能左右并排；缺了 Chrome 里经常被拆成上下两行
  const leftStyle = [
    'display:inline-block',
    'background:#fff',
    'color:#111',
    'padding:3px 8px',
    'border-radius:4px 0 0 4px',
    'font-weight:700',
    'font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace',
  ].join(';')

  const rightStyle = [
    'display:inline-block',
    'background:linear-gradient(90deg,#00c6ff 0%,#f5a623 48%,#ff2d95 100%)',
    'color:#fff',
    'padding:3px 8px',
    'border-radius:0 4px 4px 0',
    'font-weight:600',
    'font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace',
  ].join(';')

  // 清掉 Vite connecting 等更早的日志，让品牌条出现在控制台最前面
  // eslint-disable-next-line no-console -- 启动标识：清屏后置顶打印
  console.clear()
  // eslint-disable-next-line no-console -- 启动标识：双徽章 log
  console.log(`%c⚡ ${name}%c(Build: ${buildTime})`, leftStyle, rightStyle)
}
