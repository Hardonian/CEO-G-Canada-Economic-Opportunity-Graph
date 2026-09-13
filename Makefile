.PHONY: all bootstrap build test cegs-validate seed demo api worker web-build web-dev verify clean

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
	go test -v -race ./internal/... ./adapters/...

cegs-validate: build
	go test -v ./internal/cegs -run TestSpecExamples
	./bin/cog cegs validate spec/cegs/examples/project.json
	./bin/cog cegs validate spec/cegs/examples/organization.json
	./bin/cog cegs validate spec/cegs/examples/event.json
	./bin/cog cegs validate spec/cegs/examples/relationship.json
	./bin/cog cegs validate spec/cegs/examples/evidence.json
	./bin/cog cegs validate data/cegs/manifest.json

seed:
	go run ./scripts/generate_datasets.go

demo: build
	./bin/cog demo

api:
	go run ./cmd/api

worker:
	go run ./cmd/worker

web-build:
	cd apps/web && pnpm build

web-dev:
	cd apps/web && pnpm dev

verify: build test seed cegs-validate demo
	@echo "=== ALL VERIFICATION GATES PASSED ==="

clean:
	rm -rf bin
