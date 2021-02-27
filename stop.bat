@echo off

set dir=%~dp0
cd /D %dir%
log_service.exe -cmd stop -d true

pause
