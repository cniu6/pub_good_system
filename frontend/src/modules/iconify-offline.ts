/**
 * 图标离线化：把本地 @iconify-json 集合注册进 @iconify/vue/offline。
 * 之后 NovaIcon / renderIcon 按名字「按需」从内存取 SVG，不再请求 api.iconify.design。
 * 断网也能正常显示。
 */
import { addCollection } from '@iconify/vue/offline'
import iconParkOutline from '@iconify-json/icon-park-outline/icons.json'

let installed = false

export function setupIconifyOffline() {
  if (installed)
    return
  // icons.json 在构建时打进产物；运行时按图标名取用，不会再打远程。
  addCollection(iconParkOutline as Parameters<typeof addCollection>[0])
  installed = true
}
