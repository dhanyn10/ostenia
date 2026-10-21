# Run full frontend and backend test suites on Windows
$ErrorActionPreference = "Stop"

Write-Host "==> Running Go Backend Tests with Coverage..." -ForegroundColor Cyan
New-Item -Path "frontend/dist" -ItemType Directory -Force | Out-Null
"dummy" | Out-File -FilePath "frontend/dist/index.html" -Encoding utf8

# Use space-separated -coverprofile flag to prevent PowerShell dot-parsing issues
go test -v -coverprofile coverage-backend.out ./...
if ($LASTEXITCODE -ne 0) {
    Write-Error "Backend tests failed."
    exit $LASTEXITCODE
}

Write-Host "==> Running Frontend Vitest Tests with Coverage..." -ForegroundColor Cyan
Push-Location frontend
try {
    npx vitest run --coverage.enabled --coverage.reporter=lcov --coverage.reportsDirectory=./coverage
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Frontend tests failed."
        exit $LASTEXITCODE
    }
} finally {
    Pop-Location
}

Write-Host "==> All Tests Passed Successfully!" -ForegroundColor Green
