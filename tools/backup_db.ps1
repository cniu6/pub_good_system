<#
.SYNOPSIS
  按根目录 .env 的 DB_DRIVER 备份数据库（MySQL mysqldump / SQLite 文件拷贝 / Postgres pg_dump）。
.DESCRIPTION
  - 从不硬编码账号密码，全部从环境变量或 .env 读取
  - 保留最近 N 天备份（-KeepDays，默认 7）
  - 在仓库根目录执行：powershell -File tools/backup_db.ps1
#>
[CmdletBinding()]
param(
  [string]$EnvFile = "",
  [string]$BackupDir = "",
  [int]$KeepDays = 7
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
    # 已有进程环境变量优先，不覆盖
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

function Remove-OldBackups {
  param([string]$Dir, [int]$Days)
  if ($Days -lt 1) { return }
  $cutoff = (Get-Date).AddDays(-$Days)
  Get-ChildItem -LiteralPath $Dir -File -ErrorAction SilentlyContinue |
    Where-Object { $_.LastWriteTime -lt $cutoff -and ($_.Name -like "fst_backup_*") } |
    ForEach-Object {
      Write-Host "删除过期备份: $($_.FullName)"
      Remove-Item -LiteralPath $_.FullName -Force
    }
}

$root = Resolve-RepoRoot
if ([string]::IsNullOrWhiteSpace($EnvFile)) {
  $EnvFile = Join-Path $root ".env"
}
if ([string]::IsNullOrWhiteSpace($BackupDir)) {
  $BackupDir = Join-Path $root "backups"
}

Import-DotEnv -Path $EnvFile
New-Item -ItemType Directory -Force -Path $BackupDir | Out-Null

$driver = (Get-EnvOrDefault "DB_DRIVER" "mysql").ToLowerInvariant()
$stamp = Get-Date -Format "yyyyMMdd_HHmmss"
Write-Host "DB_DRIVER=$driver  BackupDir=$BackupDir  KeepDays=$KeepDays"

switch -Regex ($driver) {
  "^(mysql)?$" {
    $hostName = Get-EnvOrDefault "DB_HOST" "127.0.0.1"
    $port = Get-EnvOrDefault "DB_PORT" "3306"
    $user = Get-EnvOrDefault "DB_USER" "root"
    $pass = Get-EnvOrDefault "DB_PASSWORD" ""
    $name = Get-EnvOrDefault "DB_NAME" "fst_platform"
    if ([string]::IsNullOrWhiteSpace($pass)) {
      throw "DB_PASSWORD 为空，拒绝备份（请从 .env 或环境变量提供）"
    }
    $out = Join-Path $BackupDir "fst_backup_mysql_${name}_${stamp}.sql"
    $env:MYSQL_PWD = $pass
    try {
      & mysqldump --host=$hostName --port=$port --user=$user --single-transaction --routines --triggers --databases $name | Set-Content -LiteralPath $out -Encoding UTF8
      if ($LASTEXITCODE -ne 0) { throw "mysqldump 失败，exit=$LASTEXITCODE" }
    } finally {
      Remove-Item Env:MYSQL_PWD -ErrorAction SilentlyContinue
    }
    Write-Host "已备份 MySQL -> $out"
  }
  "^(sqlite|sqlite3)$" {
    $dbPath = Get-EnvOrDefault "DB_PATH" ""
    if ([string]::IsNullOrWhiteSpace($dbPath)) {
      $dbPath = Join-Path $root "data\fst.db"
      if (-not (Test-Path -LiteralPath $dbPath)) {
        $alt = Join-Path $root "fst.db"
        if (Test-Path -LiteralPath $alt) { $dbPath = $alt }
      }
    }
    if (-not [System.IO.Path]::IsPathRooted($dbPath)) {
      $dbPath = Join-Path $root $dbPath
    }
    if (-not (Test-Path -LiteralPath $dbPath)) {
      throw "SQLite 文件不存在: $dbPath"
    }
    $out = Join-Path $BackupDir "fst_backup_sqlite_${stamp}.db"
    Copy-Item -LiteralPath $dbPath -Destination $out -Force
    Write-Host "已备份 SQLite -> $out"
  }
  "^(postgres|postgresql|pg)$" {
    $hostName = Get-EnvOrDefault "DB_HOST" "127.0.0.1"
    $port = Get-EnvOrDefault "DB_PORT" "5432"
    $user = Get-EnvOrDefault "DB_USER" "postgres"
    $pass = Get-EnvOrDefault "DB_PASSWORD" ""
    $name = Get-EnvOrDefault "DB_NAME" "fst_platform"
    if ([string]::IsNullOrWhiteSpace($pass)) {
      throw "DB_PASSWORD 为空，拒绝备份（请从 .env 或环境变量提供）"
    }
    $out = Join-Path $BackupDir "fst_backup_postgres_${name}_${stamp}.dump"
    $env:PGPASSWORD = $pass
    try {
      & pg_dump --host=$hostName --port=$port --username=$user --format=custom --file=$out $name
      if ($LASTEXITCODE -ne 0) { throw "pg_dump 失败，exit=$LASTEXITCODE" }
    } finally {
      Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
    }
    Write-Host "已备份 PostgreSQL -> $out"
  }
  default {
    throw "不支持的 DB_DRIVER=$driver（支持 mysql / sqlite / postgres）"
  }
}

Remove-OldBackups -Dir $BackupDir -Days $KeepDays
Write-Host "备份完成"
