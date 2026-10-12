// naive-ui's light scrollbar tokens, inlined so startup does not import its full theme.
const setupScrollbarStyle = () => {
  const style = document.createElement('style')
  style.innerHTML = `
    ::-webkit-scrollbar {
      background-color: transparent;
      width: 5px;
    }
    ::-webkit-scrollbar-thumb {
      background-color: rgba(0, 0, 0, 0.25);
      border-radius: 5px;
    }
  `
  document.head.appendChild(style)
}

export default setupScrollbarStyle
