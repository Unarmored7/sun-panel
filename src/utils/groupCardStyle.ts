import { PanelPanelConfigStyleEnum } from '@/enums'

export function parseGroupCardStyle(value?: string): Panel.GroupCardStyle {
  try {
    const parsed = JSON.parse(value || '{}')
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {}
  }
  catch { return {} }
}

export function resolveGroupCardStyle(group: Panel.ItemIconGroup, config: Panel.panelConfig) {
  const style = parseGroupCardStyle(group.cardStyle)
  return {
    style: style.style === 0 || style.style === 1 ? style.style : (config.iconStyle ?? PanelPanelConfigStyleEnum.icon),
    textColor: style.textColor || config.iconTextColor,
    hideDescription: style.textInfoHideDescription ?? config.iconTextInfoHideDescription ?? false,
    hideTitle: style.textIconHideTitle ?? config.iconTextIconHideTitle ?? false,
  }
}
