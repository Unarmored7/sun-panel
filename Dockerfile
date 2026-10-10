# syntax=docker/dockerfile:1
ARG NODE_IMAGE=node:24.21.0-bookworm-slim
ARG GO_IMAGE=golang:1.27.2-bookworm
ARG RUNTIME_IMAGE=debian:bookworm-slim

FROM ${NODE_IMAGE} AS web
WORKDIR /build
RUN npm install --global pnpm@8.15.9
COPY package.json pnpm-lock.yaml .npmrc ./
RUN pnpm install --frozen-lockfile
COPY . .
RUN pnpm run build

FROM ${GO_IMAGE} AS backend
WORKDIR /build
COPY service/go.mod service/go.sum ./
RUN go mod download
COPY service/ ./
RUN go test ./... && go build -trimpath -o /sun-panel \
    -ldflags="-s -w -X sun-panel/global.RUNCODE=release -X sun-panel/global.ISDOCKER=docker" .

FROM ${RUNTIME_IMAGE} AS runtime
ARG SOURCE_REVISION=unknown
LABEL org.opencontainers.image.source="https://github.com/Unarmored7/sun-panel" \
      org.opencontainers.image.revision="${SOURCE_REVISION}" \
      org.opencontainers.image.licenses="MIT"
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=web /build/dist ./web
COPY --from=backend /sun-panel ./sun-panel
EXPOSE 3002
CMD ["./sun-panel"]
