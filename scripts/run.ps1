# Downloads and runs the latest Windows release without installing anything.
# Author: Simon Tian
param(
    [Parameter(Mandatory = $true)]
    [string]$Repository
)

$ErrorActionPreference = "Stop"
$releaseBaseUrl = "https://github.com/$Repository/releases/latest/download"
$temporaryName = "moon-$([Guid]::NewGuid().ToString('N')).exe"
$destination = Join-Path ([System.IO.Path]::GetTempPath()) $temporaryName
$checksumFile = "$destination.sha256"
try {
    Invoke-WebRequest "$releaseBaseUrl/moon-windows-amd64.exe" -OutFile $destination
    Invoke-WebRequest "$releaseBaseUrl/SHA256SUMS" -OutFile $checksumFile
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
