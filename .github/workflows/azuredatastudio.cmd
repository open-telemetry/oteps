@echo off
setlocal
set VSCODE_DEV=
set ELECTRON_RUN_AS_NODE=1
"%~dp0..\azuredatastudio.exe" "%~dp0..\resources\app\out\cli.js" %*
endlocal