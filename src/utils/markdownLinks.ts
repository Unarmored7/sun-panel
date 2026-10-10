import type { MarkdownIt } from 'markdown-it'

// Use markdown-it's own renderer types to avoid conflicting plugin type packages.
export function openExternalLinks(md: MarkdownIt) {
  const render = md.renderer.rules.link_open
  md.renderer.rules.link_open = (tokens, index, options, env, renderer) => {
    tokens[index].attrSet('target', '_blank')
    tokens[index].attrSet('rel', 'noopener noreferrer')
    return render ? render(tokens, index, options, env, renderer) : renderer.renderToken(tokens, index, options)
  }
}
