# Downloads and runs the latest Windows release without installing anything.
# Author: Simon Tian
param(
    [Parameter(Mandatory = $true)]
    [string]$Repository
)

$ErrorActionPreference = "Stop"
$release = Invoke-RestMethod "https://api.github.com/repos/$Repository/releases/latest"
$asset = $release.assets | Where-Object name -eq "moon-windows-amd64.exe" | Select-Object -First 1
$checksums = $release.assets | Where-Object name -eq "SHA256SUMS" | Select-Object -First 1
if (-not $asset -or -not $checksums) {
    throw "The latest release is missing the Windows executable or its checksums."
}

$destination = Join-Path ([System.IO.Path]::GetTempPath()) "moon-$($release.tag_name).exe"
$checksumFile = "$destination.sha256"
try {
    Invoke-WebRequest $asset.browser_download_url -OutFile $destination
    Invoke-WebRequest $checksums.browser_download_url -OutFile $checksumFile
    $checksumLine = Get-Content $checksumFile | Where-Object { $_ -match "moon-windows-amd64\.exe$" } | Select-Object -First 1
    if (-not $checksumLine) {
        throw "The checksum file does not list the Windows executable."
    }
    $expected = ($checksumLine -split "\s+")[0]
    $actual = (Get-FileHash -Algorithm SHA256 $destination).Hash
    if ($actual -ne $expected) {
        throw "The downloaded executable failed SHA-256 verification."
    }
    & $destination
} finally {
    Remove-Item -LiteralPath $destination, $checksumFile -Force -ErrorAction SilentlyContinue
}
