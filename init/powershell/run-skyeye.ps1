Push-Location $PSScriptRoot

New-Item -ItemType Directory -Path logs -Force | Out-Null
$timestamp = Get-Date -Format "yyyy-MM-dd-HH-mm-ss"
$stderrPath = "logs\$timestamp-skyeye.log"
New-Item -ItemType File -Path $stderrPath -Force | Out-Null

.\skyeye.exe `
    --config-file=config.yaml `
    --log-format=json `
    2>&1 | Tee-Object -FilePath $stderrPath -Append