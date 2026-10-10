// Bundle trusted repository SVGs directly, without an obsolete SVG build toolchain.
const icons = import.meta.glob<string>('../assets/svg-icons/*.svg', { eager: true, query: '?raw', import: 'default' })

export function setupSvgIcons() {
  const namespace = 'http://www.w3.org/2000/svg'
  const sprite = document.createElementNS(namespace, 'svg')
  sprite.setAttribute('aria-hidden', 'true')
  sprite.style.display = 'none'
  for (const [file, source] of Object.entries(icons)) {
    const root = new DOMParser().parseFromString(source, 'image/svg+xml').documentElement
    const symbol = document.createElementNS(namespace, 'symbol')
    for (const attribute of Array.from(root.attributes)) {
      if (!['width', 'height', 'xmlns', 'id'].includes(attribute.name))
        symbol.setAttribute(attribute.name, attribute.value)
    }
    symbol.id = file.split('/').pop()!.replace(/\.svg$/, '')
    for (const child of Array.from(root.childNodes))
      symbol.appendChild(document.importNode(child, true))
    sprite.appendChild(symbol)
  }
  document.body.appendChild(sprite)
}
