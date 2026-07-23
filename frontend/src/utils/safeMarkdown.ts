/**
 * 简易 Markdown HTML 消毒：去掉 script/iframe 与事件属性，
 * 仅保留常用安全标签，供公告预览等 HTML 渲染前使用。
 */

/** 允许的基础标签（小写） */
const ALLOWED_TAGS = new Set([
  'p',
  'br',
  'strong',
  'em',
  'ul',
  'ol',
  'li',
  'a',
  'code',
  'pre',
  'blockquote',
  'h1',
  'h2',
  'h3',
])

/** 事件属性名（on*） */
const ON_ATTR_RE = /^\s*on/i

/**
 * 消毒 Markdown 渲染后的 HTML 字符串。
 * 策略：DOM 解析 → 剔除危险节点/属性 → 序列化回字符串。
 */
export function sanitizeMarkdownHtml(html: string): string {
  const raw = String(html || '')
  if (!raw.trim())
    return ''

  // 无 DOM 环境时做最粗暴的字符串剥离兜底
  if (typeof document === 'undefined') {
    return raw
      .replace(/<\s*(?:script|iframe)[^>]*>[\s\S]*?<\s*\/\s*(?:script|iframe)\s*>/gi, '')
      .replace(/<\s*(?:script|iframe)[^>]*>/gi, '')
      .replace(/\son\w+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, '')
  }

  const template = document.createElement('template')
  template.innerHTML = raw

  const walk = (node: Node) => {
    // 从后往前遍历，便于安全删除
    const children = Array.from(node.childNodes)
    for (const child of children) {
      if (child.nodeType === Node.ELEMENT_NODE) {
        const el = child as HTMLElement
        const tag = el.tagName.toLowerCase()

        // 直接移除危险/未允许标签（保留文本子节点到父级）
        if (tag === 'script' || tag === 'iframe' || !ALLOWED_TAGS.has(tag)) {
          // 未允许标签：展开子节点到父级后再删自身（避免丢正文文本）
          while (el.firstChild)
            el.parentNode?.insertBefore(el.firstChild, el)
          el.remove()
          continue
        }

        // 清理属性：去掉 on*；a 仅保留安全 href
        const attrs = Array.from(el.attributes)
        for (const attr of attrs) {
          const name = attr.name.toLowerCase()
          if (ON_ATTR_RE.test(name) || name === 'style' || name.startsWith('data-')) {
            el.removeAttribute(attr.name)
            continue
          }
          if (tag === 'a' && name === 'href') {
            const href = attr.value.trim()
            // 禁止 javascript: / data: 等危险协议
            if (/^(?:javascript|data|vbscript):/i.test(href))
              el.removeAttribute(attr.name)
            else
              el.setAttribute('rel', 'noopener noreferrer')
            continue
          }
          // 非 a 的其它标签：仅保留极少数安全属性（如 code 的 class 也不保留，尽量干净）
          if (!(tag === 'a' && name === 'href') && !(tag === 'a' && name === 'title'))
            el.removeAttribute(attr.name)
        }

        walk(el)
      }
      else if (child.nodeType === Node.COMMENT_NODE) {
        child.parentNode?.removeChild(child)
      }
    }
  }

  walk(template.content)
  return template.innerHTML
}
