$ErrorActionPreference = "Stop"

Write-Host "Starting wails build..."
$logFile = "wails_build.log"

# Run wails build
# We don't need Start-Process backgrounding for build as it's a finite process
# But we redirect output to capture logs for debugging if it fails
wails build 2>&1 | Tee-Object -FilePath $logFile

if ($LASTEXITCODE -ne 0) {
    Write-Error "wails build failed with exit code $LASTEXITCODE."
    exit 1
}

$binaryPath = "build\bin\ostenia.exe"
if (Test-Path $binaryPath) {
    Write-Host "Successfully verified wails build. Binary created at $binaryPath"
    exit 0
} else {
    Write-Error "wails build finished but binary not found at $binaryPath."
    exit 1
}
