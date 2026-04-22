#Requires -Version 5.1
$ErrorActionPreference = 'Stop'

$Repo = 'Aaronmacaron/dict'
$Binary = 'dict'
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\dict' }

$arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
    'X64'   { 'amd64' }
    'Arm64' { 'arm64' }
    default { throw "unsupported architecture: $_" }
}

$release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
$version = $release.tag_name
if (-not $version) { throw 'could not determine latest version' }

$archive = "${Binary}_windows_${arch}.zip"
$baseUrl = "https://github.com/$Repo/releases/download/$version"

Write-Host "Installing $Binary $version (windows/$arch)..."

$tmp = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "dict-install-$([guid]::NewGuid())") -Force
try {
    $archivePath   = Join-Path $tmp $archive
    $checksumsPath = Join-Path $tmp 'checksums.txt'

    Invoke-WebRequest "$baseUrl/$archive"        -OutFile $archivePath   -UseBasicParsing
    Invoke-WebRequest "$baseUrl/checksums.txt"   -OutFile $checksumsPath -UseBasicParsing

    $expected = (Select-String -Path $checksumsPath -Pattern ([regex]::Escape($archive))).Line.Split(' ')[0]
    $actual   = (Get-FileHash -Algorithm SHA256 $archivePath).Hash.ToLower()
    if ($expected -ne $actual) {
        throw "checksum mismatch: expected $expected, got $actual"
    }

    Expand-Archive -Path $archivePath -DestinationPath $tmp -Force

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    Move-Item -Path (Join-Path $tmp "${Binary}.exe") -Destination (Join-Path $InstallDir "${Binary}.exe") -Force

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $onPath = ($userPath -split ';') -contains $InstallDir
    if (-not $onPath) {
        $newPath = if ([string]::IsNullOrEmpty($userPath)) { $InstallDir } else { "$userPath;$InstallDir" }
        [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
        Write-Host "Added $InstallDir to user PATH. Restart your shell to pick up the change."
    }

    Write-Host "Installed: $(Join-Path $InstallDir "${Binary}.exe")"
    Write-Host "Run '${Binary} --help' to get started."
} finally {
    Remove-Item $tmp -Recurse -Force
}
