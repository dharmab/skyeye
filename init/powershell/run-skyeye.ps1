Push-Location $PSScriptRoot

New-Item -ItemType Directory -Path logs -Force
$timestamp = Get-Date -Format "yyyy-MM-dd-HH-mm-ss"
$stderrPath = "logs\$timestamp-skyeye.log"
New-Item -ItemType File -Path $stderrPath -Force

.\skyeye.exe `
    --callsign=Focus `
    --telemetry-address=your-tacview-address:42674 `
    --telemetry-password=your-telemetry-password `
    --srs-server-address=your-srs-server:5002 `
    --srs-eam-password=your-srs-password `
    --srs-frequency=135.0 `
    --whisper-model=ggml-small.en.bin `
    --log-format=json `
    2>&1 | Tee-Object -FilePath $stderrPath -Append