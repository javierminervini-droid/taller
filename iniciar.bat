@echo off
cd /d "%~dp0"
call npm install
call npm run build --prefix client
echo.
echo Abrí en el navegador: http://127.0.0.1:3847
echo Usuario: admin / admin123
echo.
call npm start
