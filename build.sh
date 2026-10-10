#!/bin/sh
# Build the maintained source. Assets are embedded by Go itself.
set -eu
cd "$(dirname "$0")"
pnpm install --frozen-lockfile
pnpm run build
(cd service && go test ./... && go build -trimpath -ldflags="-s -w -X sun-panel/global.RUNCODE=release" -o sun-panel .)
printf '%s\n' 'Built dist/ and service/sun-panel. Dockerfile creates a self-contained image.'
