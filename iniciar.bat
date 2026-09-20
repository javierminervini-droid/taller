@echo off
cd /d "%~dp0"
if not exist .env copy .env.example .env
echo.
echo Starting Postgres + API with Docker Compose...
echo If Docker is not installed, see docs\dev-setup.md
echo.
docker compose up -d --build
echo.
echo API health: http://127.0.0.1:3847/health
echo Demo users: admin/admin123  coord/coord123  diego/diego123
echo.
echo For the React UI in another window:
echo   npm run dev:client
echo Then open http://127.0.0.1:5180
echo.
pause
