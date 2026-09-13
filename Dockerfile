# syntax=docker/dockerfile:1.7

# Build the dependency-free Go services with the same toolchain declared in go.mod.
FROM golang:1.26-alpine AS go-builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /out/cog ./cmd/cog \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /out/api ./cmd/api \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /out/worker ./cmd/worker

# Build a minimal, reproducible Next.js standalone deployment.
FROM node:22-alpine AS web-builder
WORKDIR /src/apps/web
RUN corepack enable && corepack prepare pnpm@11.8.0 --activate
COPY apps/web/package.json apps/web/pnpm-lock.yaml apps/web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY apps/web/ ./
RUN pnpm build

# API/worker runtime. Kept as "runner" for backwards-compatible compose targets.
FROM alpine:3.23 AS runner
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 cog \
    && adduser -S -D -H -u 10001 -G cog cog
COPY --from=go-builder --chown=cog:cog /out/api /app/bin/api
COPY --from=go-builder --chown=cog:cog /out/worker /app/bin/worker
COPY --from=go-builder --chown=cog:cog /out/cog /app/bin/cog
COPY --chown=cog:cog data/ /app/data/
USER 10001:10001
EXPOSE 8080
ENV PORT=8080 ENV=production
ENTRYPOINT ["/app/bin/api"]

# Web runtime is isolated from build tooling and runs without root privileges.
FROM node:22-alpine AS web-runner
WORKDIR /app
ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1 HOSTNAME=0.0.0.0 PORT=3000
RUN addgroup -S -g 10001 nextjs \
    && adduser -S -D -H -u 10001 -G nextjs nextjs
COPY --from=web-builder --chown=nextjs:nextjs /src/apps/web/.next/standalone ./
COPY --from=web-builder --chown=nextjs:nextjs /src/apps/web/.next/static ./.next/static
USER 10001:10001
EXPOSE 3000
CMD ["node", "server.js"]
