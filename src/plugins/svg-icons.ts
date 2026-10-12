// The sprite is a separate chunk so its SVG sources stay out of the startup bundle;
// <use> references resolve as soon as it is mounted.
export function setupSvgIcons() {
  import('./svg-sprite').then(module => module.mountSvgSprite())
}
