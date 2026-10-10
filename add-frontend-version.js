const fs = require('fs')
const version = fs.readFileSync('service/assets/version', 'utf8').trim().split('|')[1]
const envFilePath = '.env'
let content = fs.readFileSync(envFilePath, 'utf8')
content = content.replace(/\n*VITE_APP_VERSION=.*(?:\n|$)/g, '\n').trimEnd()
fs.writeFileSync(envFilePath, `${content}\nVITE_APP_VERSION=${version}\n`)
console.log(`Frontend version: ${version}`)
