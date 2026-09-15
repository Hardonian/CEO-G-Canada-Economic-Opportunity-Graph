.PHONY: all bootstrap build test bench vet lint cegs-validate release-check seed demo api worker web-build web-dev verify clean

all: build test

bootstrap:
	go mod download
	cd apps/web && pnpm install

build:
	mkdir -p bin
	go build -o bin/cog ./cmd/cog
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

test:
	go test -v -race ./...

bench:
	go test -bench=. -benchmem ./...

vet:
	go vet ./...

lint: vet
	@which staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed; skipping"

cegs-validate: build
	go test -v ./internal/cegs -run TestSpecExamples
	./bin/cog cegs validate spec/cegs/examples/project.json
	./bin/cog cegs validate spec/cegs/examples/organization.json
	./bin/cog cegs validate spec/cegs/examples/event.json
	./bin/cog cegs validate spec/cegs/examples/relationship.json
	./bin/cog cegs validate spec/cegs/examples/evidence.json
	./bin/cog cegs validate data/cegs/manifest.json

release-check:
	go run ./cmd/releasecheck

seed:
	go run ./scripts/generate_datasets.go
	cd apps/web && pnpm snapshot:generate

demo: build
	./bin/cog demo

api:
	go run ./cmd/api

worker:
	go run ./cmd/worker

web-build:
	cd apps/web && pnpm typecheck && pnpm build

web-dev:
	cd apps/web && pnpm dev

verify: build vet test bench release-check cegs-validate demo web-build
	@echo "=== ALL VERIFICATION GATES PASSED ==="

clean:
	rm -rf bin
