# Multi-Stage Production Containerfile for CanadaOpportunityGraph

# Stage 1: Build Go Backend Binaries
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/cog ./cmd/cog
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/worker ./cmd/worker

# Stage 2: Build Next.js Web Frontend
FROM node:22-alpine AS web-builder
WORKDIR /app/web
RUN corepack enable && corepack prepare pnpm@latest --activate
COPY apps/web/package.json apps/web/pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile
COPY apps/web/ ./
RUN pnpm build

# Stage 3: Minimal Production Image
FROM alpine:3.20 AS runner
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=go-builder /app/bin/api /app/bin/api
COPY --from=go-builder /app/bin/worker /app/bin/worker
COPY --from=go-builder /app/bin/cog /app/bin/cog
COPY data/ /app/data/

EXPOSE 8080
ENV PORT=8080
ENV ENV=production

CMD ["/app/bin/api"]
