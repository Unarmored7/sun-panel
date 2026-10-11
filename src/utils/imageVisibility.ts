const callbacks = new Map<Element, () => void>()
let observer: IntersectionObserver | undefined

// A shared observer avoids starting requests for every icon in a long page.
export function loadNearViewport(element: Element, load: () => void) {
  if (typeof IntersectionObserver === 'undefined') {
    load()
    return () => {}
  }
  observer ||= new IntersectionObserver((entries) => {
    for (const entry of entries) {
      if (entry.isIntersecting) {
        callbacks.get(entry.target)?.()
        callbacks.delete(entry.target)
        observer?.unobserve(entry.target)
      }
    }
  }, { rootMargin: '250px' })
  callbacks.set(element, load)
  observer.observe(element)
  return () => { callbacks.delete(element); observer?.unobserve(element) }
}
