@echo off
REM Nova Crew Server - Web Interface Launcher (Windows)
REM This script starts a simple HTTP server to serve the web interface

echo.
echo Starting Nova Crew Server Web Interface...
echo.
echo Make sure the Go server is running on http://localhost:8080
echo If not, run: cd .. ^&^& go run main.go
echo.
echo Starting web server on http://localhost:3000
echo Press Ctrl+C to stop
echo.

REM Try Python3 first
where python3 >nul 2>nul
if %ERRORLEVEL% == 0 (
    echo Using Python 3
    python3 -m http.server 3000
    goto :end
)

REM Try Python
where python >nul 2>nul
if %ERRORLEVEL% == 0 (
    echo Using Python
    python -m http.server 3000
    goto :end
)

REM Try PHP
where php >nul 2>nul
if %ERRORLEVEL% == 0 (
    echo Using PHP
    php -S localhost:3000
    goto :end
)

REM Try Node.js npx
where npx >nul 2>nul
if %ERRORLEVEL% == 0 (
    echo Using Node.js http-server
    npx http-server -p 3000
    goto :end
)

REM No server found
echo Error: No suitable HTTP server found
echo.
echo Please install one of the following:
echo   - Python 3: https://www.python.org/
echo   - Node.js: https://nodejs.org/
echo   - PHP: https://www.php.net/
pause
exit /b 1

:end
