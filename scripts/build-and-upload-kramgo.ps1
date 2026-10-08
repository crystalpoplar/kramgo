param(
    [string]$RemoteHost = "dtk1376@192.168.1.49",
    [string]$RemoteDir = "/home/dtk1376",
    [string]$Goos = "linux",
    [string]$Goarch = "amd64"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$distDir = Join-Path $repoRoot "dist\linux-amd64"
New-Item -ItemType Directory -Force -Path $distDir | Out-Null

$env:GOOS = $Goos
$env:GOARCH = $Goarch
$env:CGO_ENABLED = "0"

Write-Host "Building kramgo for $Goos/$Goarch..."
go build -o (Join-Path $distDir "kramgo") "$repoRoot/cmd/kramgo"
go build -o (Join-Path $distDir "kramgo-watchdog") "$repoRoot/cmd/kramgo-watchdog"

Copy-Item (Join-Path $repoRoot "kramgo-ollama.service") $distDir -Force
Copy-Item (Join-Path $repoRoot "kramgo-watchdog.service") $distDir -Force
Copy-Item (Join-Path $repoRoot "kramgo-reboot.service") $distDir -Force
Copy-Item (Join-Path $repoRoot "kramgo-reboot.timer") $distDir -Force
Copy-Item (Join-Path $repoRoot "install-kramgo-ollama.sh") $distDir -Force

Write-Host "Uploading to ${RemoteHost}:${RemoteDir}"
ssh -o StrictHostKeyChecking=no "${RemoteHost}" "mkdir -p '${RemoteDir}'"

$files = @(
    (Join-Path $distDir "kramgo"),
    (Join-Path $distDir "kramgo-watchdog"),
    (Join-Path $distDir "kramgo-ollama.service"),
    (Join-Path $distDir "kramgo-watchdog.service"),
    (Join-Path $distDir "kramgo-reboot.service"),
    (Join-Path $distDir "kramgo-reboot.timer"),
    (Join-Path $distDir "install-kramgo-ollama.sh")
)

foreach ($file in $files) {
    scp -o StrictHostKeyChecking=no "$file" "${RemoteHost}:${RemoteDir}/"
}

$remoteFix = @"
for f in \
  '${RemoteDir}/install-kramgo-ollama.sh' \
  '${RemoteDir}/kramgo-ollama.service' \
  '${RemoteDir}/kramgo-watchdog.service' \
  '${RemoteDir}/kramgo-reboot.service' \
  '${RemoteDir}/kramgo-reboot.timer'; do
  sed -i 's/\r$//' "$f"
done
chmod 755 '${RemoteDir}/kramgo' '${RemoteDir}/kramgo-watchdog' '${RemoteDir}/install-kramgo-ollama.sh'
"@

ssh -o StrictHostKeyChecking=no "${RemoteHost}" "$remoteFix"

Write-Host "Upload complete. Run this on the Ubuntu host:"
Write-Host "  sudo ${RemoteDir}/install-kramgo-ollama.sh"
