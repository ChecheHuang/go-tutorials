@echo off
setlocal

if "%1"=="" goto run
if "%1"=="run" goto run
if "%1"=="build" goto build
if "%1"=="test" goto test
if "%1"=="cover" goto cover
if "%1"=="clean" goto clean
if "%1"=="swagger" goto swagger
if "%1"=="tidy" goto tidy
goto help

:run
go run ./cmd/server/
goto end

:build
go build -o blog-api.exe ./cmd/server/
echo. && echo [OK] blog-api.exe
goto end

:test
go test -v ./internal/...
goto end

:cover
go test -cover ./internal/...
goto end

:clean
del /f blog-api.exe blog.db 2>nul
echo [OK] clean
goto end

:swagger
swag init -g cmd/server/main.go -o docs
goto end

:tidy
go mod tidy
goto end

:help
echo.
echo   run        run run           start server (default)
echo   run build  run build         compile to blog-api.exe
echo   run test   run test          run tests
echo   run cover  run cover         test coverage
echo   run clean  run clean         delete db and binary
echo   run swagger                  regenerate swagger docs
echo   run tidy   run tidy          go mod tidy
echo.
goto end

:end
endlocal
