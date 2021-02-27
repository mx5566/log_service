@echo off

set dir=%~dp0
cd /D %dir%

log_service.exe -cmd uninstall -d true

pause