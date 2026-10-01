[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [string] $Command,

    [Parameter(Position = 1, ValueFromRemainingArguments = $true)]
    [string[]] $Arguments
)

Set-StrictMode -Version 2.0
$ErrorActionPreference = 'Stop'
$Arguments = @($Arguments)
if ($null -eq $Arguments) {
    $ArgumentCount = 0
} else {
    $ArgumentCount = $Arguments.Count
}

$ApplicationName = 'Tualo File Opener'
$DefaultScheme = 'tualo-fs'
$ProgramId = 'TualoFileOpener.URL'
$CapabilitiesPath = 'HKCU:\Software\Tualo\FileOpener\Capabilities'
$ConfigPath = Join-Path ([Environment]::GetFolderPath('ApplicationData')) 'tualo-fileopener\config.json'

function Show-Usage {
    @'
Tualo File Opener

Verwendung:
  fileopener.ps1 install [Schema]
  fileopener.ps1 set <Alias> <Ordner>
  fileopener.ps1 list
  fileopener.ps1 open <Schema://Alias>
  fileopener.ps1 remove <Alias>
  fileopener.ps1 uninstall
'@ | Write-Host
}

function Get-Configuration {
    if (-not (Test-Path -LiteralPath $ConfigPath)) {
        return [pscustomobject]@{
            folders = [pscustomobject]@{}
            scheme = $DefaultScheme
        }
    }

    $configuration = Get-Content -LiteralPath $ConfigPath -Raw | ConvertFrom-Json
    if ($null -eq $configuration.folders) {
        $configuration | Add-Member -NotePropertyName folders -NotePropertyValue ([pscustomobject]@{})
    }
    if ([string]::IsNullOrEmpty([string]$configuration.scheme)) {
        $configuration | Add-Member -NotePropertyName scheme -NotePropertyValue $DefaultScheme
    }
    $configuration
}

function Save-Configuration([object] $Configuration) {
    $directory = Split-Path -Parent $ConfigPath
    New-Item -ItemType Directory -Path $directory -Force | Out-Null
    $Configuration | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $ConfigPath -Encoding UTF8
}

function Normalize-Scheme([string] $Scheme) {
    if ([string]::IsNullOrEmpty($Scheme) -or $Scheme -notmatch '^[A-Za-z][A-Za-z0-9+.-]*$') {
        throw 'Schema muss mit einem Buchstaben beginnen und darf danach nur Buchstaben, Ziffern sowie +, - oder . enthalten.'
    }
    $Scheme.ToLowerInvariant()
}

function Assert-Alias([string] $Alias) {
    if ([string]::IsNullOrEmpty($Alias) -or $Alias.Trim() -ne $Alias -or $Alias -match '[\\/]') {
        throw 'Alias darf nicht leer sein und keine Schraegstriche enthalten.'
    }
}

function Get-FolderEntries([object] $Configuration) {
    if ($null -eq $Configuration.folders) {
        return @()
    }
    @($Configuration.folders.psobject.Properties)
}

function Find-Folder([object] $Configuration, [string] $Alias) {
    foreach ($entry in (Get-FolderEntries $Configuration)) {
        if ($entry.Name -ieq $Alias) {
            $folder = Get-Item -LiteralPath ([string]$entry.Value) -Force
            if (-not $folder.PSIsContainer) {
                throw "Pfad fuer Alias '$($entry.Name)' ist kein Ordner."
            }
            return $folder.FullName
        }
    }
    throw "Alias '$Alias' ist nicht konfiguriert."
}

function Get-FolderForTarget([object] $Configuration, [string] $Alias, [string] $RelativePath) {
    $root = Find-Folder $Configuration $Alias
    if ([string]::IsNullOrEmpty($RelativePath)) {
        return $root
    }

    $segments = $RelativePath.Trim('/').Split('/')
    foreach ($segment in $segments) {
        if ([string]::IsNullOrEmpty($segment) -or $segment -eq '.' -or $segment -eq '..' -or $segment.Contains('\')) {
            throw 'Link enthaelt einen ungueltigen Unterpfad.'
        }
    }

    $target = Get-Item -LiteralPath (Join-Path $root ($segments -join '\')) -Force
    if (-not $target.PSIsContainer) {
        throw "Unterpfad '$RelativePath' ist kein Ordner."
    }
    $rootPath = (Get-Item -LiteralPath $root -Force).FullName.TrimEnd('\') + '\'
    if (-not $target.FullName.StartsWith($rootPath, [StringComparison]::OrdinalIgnoreCase)) {
        throw 'Unterordner liegt ausserhalb des konfigurierten Stammordners.'
    }
    $target.FullName
}

function Get-TargetFromUrl([string] $RawUrl, [string] $ExpectedScheme) {
    try {
        $uri = [Uri]$RawUrl
    } catch {
        throw 'Link ist ungueltig.'
    }
    if (-not $uri.Scheme.Equals($ExpectedScheme, [StringComparison]::OrdinalIgnoreCase)) {
        throw "Schema muss '$ExpectedScheme' sein."
    }
    if ($uri.UserInfo -or $uri.Query -or $uri.Fragment -or $uri.Port -gt 0) {
        throw 'Link darf keine Zugangsdaten, Parameter, Fragmente oder Ports enthalten.'
    }

    $alias = if ($uri.Host) { $uri.Host } else { $uri.OriginalString.Substring($uri.Scheme.Length + 1).Split('/')[0] }
    $alias = [Uri]::UnescapeDataString($alias)
    Assert-Alias $alias
    $relativePath = [Uri]::UnescapeDataString($uri.AbsolutePath).Trim('/')
    [pscustomobject]@{ Alias = $alias; RelativePath = $relativePath }
}

function Open-ConfiguredUrl([string] $RawUrl) {
    $configuration = Get-Configuration
    $target = Get-TargetFromUrl $RawUrl ([string]$configuration.scheme)
    $folder = Get-FolderForTarget $configuration $target.Alias $target.RelativePath
    Start-Process -FilePath 'explorer.exe' -ArgumentList ('"{0}"' -f $folder)
}

function Set-Folder([object] $Configuration, [string] $Alias, [string] $Path) {
    Assert-Alias $Alias
    $folder = Get-Item -LiteralPath $Path -Force
    if (-not $folder.PSIsContainer) {
        throw "'$Path' ist kein Ordner."
    }
    foreach ($entry in @(Get-FolderEntries $Configuration)) {
        if ($entry.Name -ieq $Alias) {
            $Configuration.folders.psobject.Properties.Remove($entry.Name)
        }
    }
    $Configuration.folders | Add-Member -NotePropertyName $Alias -NotePropertyValue $folder.FullName
}

function Install-Protocol([string] $Scheme) {
    $scriptPath = $PSCommandPath
    $command = 'powershell.exe -NoProfile -ExecutionPolicy Bypass -File "{0}" open "%1"' -f $scriptPath
    $schemeKey = "HKCU:\Software\Classes\$Scheme"
    $programKey = "HKCU:\Software\Classes\$ProgramId"

    New-Item -Path "$schemeKey\shell\open\command" -Force | Out-Null
    Set-ItemProperty -Path $schemeKey -Name '(default)' -Value "URL:$ApplicationName"
    New-ItemProperty -Path $schemeKey -Name 'URL Protocol' -Value '' -PropertyType String -Force | Out-Null
    Set-ItemProperty -Path "$schemeKey\shell\open\command" -Name '(default)' -Value $command

    New-Item -Path "$programKey\shell\open\command" -Force | Out-Null
    Set-ItemProperty -Path $programKey -Name '(default)' -Value "URL:$ApplicationName"
    New-ItemProperty -Path $programKey -Name 'URL Protocol' -Value '' -PropertyType String -Force | Out-Null
    Set-ItemProperty -Path "$programKey\shell\open\command" -Name '(default)' -Value $command

    New-Item -Path "$CapabilitiesPath\URLAssociations" -Force | Out-Null
    Set-ItemProperty -Path $CapabilitiesPath -Name ApplicationName -Value $ApplicationName
    Set-ItemProperty -Path $CapabilitiesPath -Name ApplicationDescription -Value 'Oeffnet konfigurierte lokale Ordner ueber URL-Links.'
    Set-ItemProperty -Path "$CapabilitiesPath\URLAssociations" -Name $Scheme -Value $ProgramId
    New-Item -Path 'HKCU:\Software\RegisteredApplications' -Force | Out-Null
    Set-ItemProperty -Path 'HKCU:\Software\RegisteredApplications' -Name $ApplicationName -Value 'Software\Tualo\FileOpener\Capabilities'
}

function Uninstall-Protocol([string] $Scheme) {
    Remove-ItemProperty -Path "$CapabilitiesPath\URLAssociations" -Name $Scheme -ErrorAction SilentlyContinue
    Remove-ItemProperty -Path 'HKCU:\Software\RegisteredApplications' -Name $ApplicationName -ErrorAction SilentlyContinue
    Remove-Item -Path $CapabilitiesPath -Recurse -Force -ErrorAction SilentlyContinue
    Remove-Item -Path "HKCU:\Software\Classes\$ProgramId" -Recurse -Force -ErrorAction SilentlyContinue
    Remove-Item -Path "HKCU:\Software\Classes\$Scheme" -Recurse -Force -ErrorAction SilentlyContinue
}

try {
    if ([string]::IsNullOrEmpty($Command) -or $Command -in @('help', '-h', '--help')) {
        Show-Usage
        exit 0
    }

    $configuration = Get-Configuration
    switch ($Command.ToLowerInvariant()) {
        'install' {
            if ($ArgumentCount -gt 1) { throw 'Verwendung: install [Schema]' }
            $scheme = [string]$configuration.scheme
            if ($ArgumentCount -eq 1) { $scheme = Normalize-Scheme $Arguments[0] }
            Uninstall-Protocol $scheme
            Install-Protocol $scheme
            $configuration.scheme = $scheme
            Save-Configuration $configuration
            Write-Host "Schema ${scheme}:// wurde fuer diesen Benutzer registriert."
        }
        'uninstall' {
            if ($ArgumentCount -ne 0) { throw 'Verwendung: uninstall' }
            Uninstall-Protocol ([string]$configuration.scheme)
            Write-Host "Registrierung fuer $($configuration.scheme):// wurde entfernt."
        }
        'set' {
            if ($ArgumentCount -ne 2) { throw 'Verwendung: set <Alias> <Ordner>' }
            Set-Folder $configuration $Arguments[0] $Arguments[1]
            Save-Configuration $configuration
            Write-Host "$($Arguments[0]) -> $($Arguments[1])"
        }
        'list' {
            if ($ArgumentCount -ne 0) { throw 'Verwendung: list' }
            foreach ($entry in (Get-FolderEntries $configuration | Sort-Object Name)) {
                Write-Host "$($entry.Name) -> $($entry.Value)"
            }
        }
        'remove' {
            if ($ArgumentCount -ne 1) { throw 'Verwendung: remove <Alias>' }
            $found = $false
            foreach ($entry in @(Get-FolderEntries $configuration)) {
                if ($entry.Name -ieq $Arguments[0]) {
                    $configuration.folders.psobject.Properties.Remove($entry.Name)
                    $found = $true
                }
            }
            if (-not $found) { throw "Alias '$($Arguments[0])' ist nicht konfiguriert." }
            Save-Configuration $configuration
        }
        'open' {
            if ($ArgumentCount -ne 1) { throw 'Verwendung: open <Schema://Alias>' }
            Open-ConfiguredUrl $Arguments[0]
        }
        default {
            if ($ArgumentCount -eq 0) { Open-ConfiguredUrl $Command } else { throw "Unbekannter Befehl '$Command'." }
        }
    }
} catch {
    [Console]::Error.WriteLine("Fehler: $($_.Exception.Message)")
    exit 1
}