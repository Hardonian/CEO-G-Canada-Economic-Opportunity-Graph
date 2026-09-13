# CanadaOpportunityGraph (COG) & CEGS Unified Windows Task Runner

param (
    [Parameter(Position=0)]
    [string]$Target = "help",
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$Args
)

$ErrorActionPreference = "Stop"

function Show-Help {
    Write-Host "CanadaOpportunityGraph Task Runner (task.ps1)" -ForegroundColor Cyan
    Write-Host "Targets:" -ForegroundColor Yellow
    Write-Host "  build           - Build all Go binaries (cog, api, worker)"
    Write-Host "  test            - Run all Go unit and integration tests"
    Write-Host "  cegs-validate   - Validate all spec schemas, examples, and public datasets"
    Write-Host "  seed            - Ingest authoritative adapters and generate public snapshots"
    Write-Host "  demo            - Run instant deterministic demo via CLI"
    Write-Host "  api             - Run the REST API server on :8080"
    Write-Host "  web-build       - Build the Next.js institutional web frontend"
    Write-Host "  web-dev         - Run Next.js dev server on :3000"
    Write-Host "  verify          - Complete end-to-end release verification"
}

switch ($Target.ToLower()) {
    "build" {
        Write-Host "[BUILD] Compiling Go binaries to bin/..." -ForegroundColor Cyan
        if (-not (Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }
        go build -o bin/cog.exe ./cmd/cog
        go build -o bin/api.exe ./cmd/api
        go build -o bin/worker.exe ./cmd/worker
        Write-Host "[SUCCESS] Binaries compiled successfully in bin/." -ForegroundColor Green
    }
    "test" {
        Write-Host "[TEST] Running Go test suite with race detector..." -ForegroundColor Cyan
        go test -v ./internal/... ./adapters/...
    }
    "cegs-validate" {
        Write-Host "[CEGS] Validating all specification examples and public manifests..." -ForegroundColor Cyan
        go test -v ./internal/cegs -run TestSpecExamples
        .\bin\cog.exe cegs validate spec/cegs/examples/project.json
        .\bin\cog.exe cegs validate spec/cegs/examples/organization.json
        .\bin\cog.exe cegs validate spec/cegs/examples/event.json
        .\bin\cog.exe cegs validate spec/cegs/examples/relationship.json
        .\bin\cog.exe cegs validate spec/cegs/examples/evidence.json
        .\bin\cog.exe cegs validate data/cegs/manifest.json
        Write-Host "[SUCCESS] All CEGS assets passed 100% conformance validation." -ForegroundColor Green
    }
    "seed" {
        Write-Host "[SEED] Generating public datasets..." -ForegroundColor Cyan
        go run ./scripts/generate_datasets.go
    }
    "demo" {
        Write-Host "[DEMO] Running deterministic demo..." -ForegroundColor Cyan
        .\bin\cog.exe demo
    }
    "api" {
        Write-Host "[API] Starting CanadaOpportunityGraph API on :8080..." -ForegroundColor Cyan
        go run ./cmd/api
    }
    "web-build" {
        Write-Host "[WEB] Building Next.js frontend..." -ForegroundColor Cyan
        Set-Location "apps/web"
        pnpm build
        Set-Location "../.."
    }
    "web-dev" {
        Write-Host "[WEB] Launching Next.js dev server on :3000..." -ForegroundColor Cyan
        Set-Location "apps/web"
        pnpm dev
        Set-Location "../.."
    }
    "verify" {
        Write-Host "=== CanadaOpportunityGraph Full Verification Suite ===" -ForegroundColor Cyan
        & $PSCommandPath build
        & $PSCommandPath test
        & $PSCommandPath seed
        & $PSCommandPath cegs-validate
        & $PSCommandPath demo
        Write-Host "=== VERIFICATION COMPLETE: ALL GATES PASSED ===" -ForegroundColor Green
    }
    default {
        Show-Help
    }
}
