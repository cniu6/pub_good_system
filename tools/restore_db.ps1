<#
.SYNOPSIS
  从备份恢复数据库（按 DB_DRIVER：mysql / sqlite / postgres）。
.DESCRIPTION
  - 凭据只从环境变量或 .env 读取，从不硬编码
  - 恢复前强烈建议先再跑一次 backup_db.ps1
  - 用法：powershell -File tools/restore_db.ps1 -BackupFile backups\xxx.sql
#>
[CmdletBinding()]
param(
  [Parameter(Mandatory = $true)]
  [string]$BackupFile,
  [string]$EnvFile = "",
  [switch]$Force
)

$ErrorActionPreference = "Stop"
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

function Resolve-RepoRoot {
  Split-Path -Parent $PSScriptRoot
}

function Import-DotEnv {
  param([string]$Path)
  if (-not (Test-Path -LiteralPath $Path)) { return }
  Get-Content -LiteralPath $Path -Encoding UTF8 | ForEach-Object {
    $line = $_.Trim()
    if ($line -eq "" -or $line.StartsWith("#")) { return }
    $idx = $line.IndexOf("=")
    if ($idx -lt 1) { return }
    $key = $line.Substring(0, $idx).Trim()
    $val = $line.Substring($idx + 1).Trim()
    if (($val.StartsWith('"') -and $val.EndsWith('"')) -or ($val.StartsWith("'") -and $val.EndsWith("'"))) {
      $val = $val.Substring(1, $val.Length - 2)
    }
    if (-not [string]::IsNullOrEmpty([Environment]::GetEnvironmentVariable($key))) { return }
    [Environment]::SetEnvironmentVariable($key, $val, "Process")
  }
}

function Get-EnvOrDefault {
  param([string]$Name, [string]$Default = "")
  $v = [Environment]::GetEnvironmentVariable($Name)
  if ([string]::IsNullOrWhiteSpace($v)) { return $Default }
  return $v.Trim()
}

$root = Resolve-RepoRoot
if ([string]::IsNullOrWhiteSpace($EnvFile)) {
  $EnvFile = Join-Path $root ".env"
}
Import-DotEnv -Path $EnvFile

if (-not (Test-Path -LiteralPath $BackupFile)) {
  # 允许相对仓库根的路径
  $alt = Join-Path $root $BackupFile
  if (Test-Path -LiteralPath $alt) { $BackupFile = $alt }
}
if (-not (Test-Path -LiteralPath $BackupFile)) {
  throw "备份文件不存在: $BackupFile"
}

$driver = (Get-EnvOrDefault "DB_DRIVER" "mysql").ToLowerInvariant()
Write-Host "DB_DRIVER=$driver  BackupFile=$BackupFile"

if (-not $Force) {
  $ans = Read-Host "将覆盖目标库数据，确认继续？(yes/no)"
  if ($ans -ne "yes") {
    Write-Host "已取消"
    exit 1
  }
}

switch -Regex ($driver) {
  "^(mysql)?$" {
    $hostName = Get-EnvOrDefault "DB_HOST" "127.0.0.1"
    $port = Get-EnvOrDefault "DB_PORT" "3306"
    $user = Get-EnvOrDefault "DB_USER" "root"
    $pass = Get-EnvOrDefault "DB_PASSWORD" ""
    if ([string]::IsNullOrWhiteSpace($pass)) {
      throw "DB_PASSWORD 为空，拒绝恢复"
    }
    $env:MYSQL_PWD = $pass
    try {
      Get-Content -LiteralPath $BackupFile -Raw -Encoding UTF8 | & mysql --host=$hostName --port=$port --user=$user
      if ($LASTEXITCODE -ne 0) { throw "mysql 恢复失败，exit=$LASTEXITCODE" }
    } finally {
      Remove-Item Env:MYSQL_PWD -ErrorAction SilentlyContinue
    }
    Write-Host "MySQL 恢复完成"
  }
  "^(sqlite|sqlite3)$" {
    $dbPath = Get-EnvOrDefault "DB_PATH" ""
    if ([string]::IsNullOrWhiteSpace($dbPath)) {
      $dbPath = Join-Path $root "data\fst.db"
    }
    if (-not [System.IO.Path]::IsPathRooted($dbPath)) {
      $dbPath = Join-Path $root $dbPath
    }
    $parent = Split-Path -Parent $dbPath
    if ($parent) { New-Item -ItemType Directory -Force -Path $parent | Out-Null }
    Copy-Item -LiteralPath $BackupFile -Destination $dbPath -Force
    Write-Host "SQLite 已恢复到 $dbPath"
  }
  "^(postgres|postgresql|pg)$" {
    $hostName = Get-EnvOrDefault "DB_HOST" "127.0.0.1"
    $port = Get-EnvOrDefault "DB_PORT" "5432"
    $user = Get-EnvOrDefault "DB_USER" "postgres"
    $pass = Get-EnvOrDefault "DB_PASSWORD" ""
    $name = Get-EnvOrDefault "DB_NAME" "fst_platform"
    if ([string]::IsNullOrWhiteSpace($pass)) {
      throw "DB_PASSWORD 为空，拒绝恢复"
    }
    $env:PGPASSWORD = $pass
    try {
      & pg_restore --host=$hostName --port=$port --username=$user --dbname=$name --clean --if-exists $BackupFile
      if ($LASTEXITCODE -ne 0) {
        # 部分警告也会非 0；若文件是纯 SQL 则尝试 psql
        Write-Host "pg_restore 返回 $LASTEXITCODE，尝试按 SQL 用 psql 导入..."
        Get-Content -LiteralPath $BackupFile -Raw | & psql --host=$hostName --port=$port --username=$user --dbname=$name
        if ($LASTEXITCODE -ne 0) { throw "PostgreSQL 恢复失败" }
      }
    } finally {
      Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
    }
    Write-Host "PostgreSQL 恢复完成"
  }
  default {
    throw "不支持的 DB_DRIVER=$driver"
  }
}
