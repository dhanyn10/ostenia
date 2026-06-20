$ErrorActionPreference = "Stop"

Write-Host "Starting wails dev..."
$logFile = "wails_dev.log"
$errFile = "wails_dev_err.log"

# Start wails dev in the background
$process = Start-Process -FilePath "wails" -ArgumentList "dev" -RedirectStandardOutput $logFile -RedirectStandardError $errFile -PassThru -WindowStyle Hidden

$timeout = 600 # 10 minutes timeout
$startTime = Get-Date
$found = $false

Write-Host "Monitoring output for success message..."

try {
    while (((Get-Date) - $startTime).TotalSeconds -lt $timeout) {
        if (Test-Path $logFile) {
            # Use -Tail to avoid reading the whole file every time if it gets large,
            # but -Raw and -match is simpler for small logs.
            $content = Get-Content $logFile
            foreach ($line in $content) {
                if ($line -like "*Serving assets from frontend DevServer URL*") {
                    Write-Host "Matched success line: $line"
                    $found = $true
                    break
                }
            }
        }

        if ($found) { break }

        if ($process.HasExited) {
            Write-Host "wails dev process exited unexpectedly with code $($process.ExitCode)."
            if (Test-Path $errFile) {
                Write-Host "Standard Error Output:"
                Get-Content $errFile
            }
            if (Test-Path $logFile) {
                Write-Host "Standard Output:"
                Get-Content $logFile
            }
            exit 1
        }

        Write-Host "Waiting... ($([int]((Get-Date) - $startTime).TotalSeconds)s)"
        Start-Sleep -Seconds 10
    }
}
finally {
    if (-not $process.HasExited) {
        Write-Host "Stopping wails dev process..."
        Stop-Process -Id $process.Id -Force
    }
}

if ($found) {
    Write-Host "Successfully verified wails dev is running."
    exit 0
} else {
    Write-Error "Timeout reached: Could not find the success message in wails dev output."
    if (Test-Path $logFile) {
        Write-Host "Full wails dev output:"
        Get-Content $logFile
    }
    exit 1
}
