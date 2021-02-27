@echo on
set dir=%~dp0
cd /D %dir%
log_service.exe -cmd start1 -d true
pause

