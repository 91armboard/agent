@echo off
setlocal EnableDelayedExpansion

set "md5="

for /f "tokens=*" %%a in ('CertUtil -hashfile main MD5 ^| findstr /i /v "MD5 hash of"') do (
  set "line=%%a"
  set "line=!line:~0,-2!"  // Remove both the carriage return and line feed characters
  set "md5=!md5!!line!"
)

echo !md5!> md5.upgrade
