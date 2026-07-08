@echo off
cd /d "%~dp0"
start "blog backend" powershell -NoProfile -ExecutionPolicy Bypass -NoExit -File "%~dp0scripts\start.ps1"
start "blog admin" powershell -NoProfile -ExecutionPolicy Bypass -NoExit -File "%~dp0scripts\start.ps1" -Admin