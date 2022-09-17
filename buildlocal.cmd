@echo off

pushd "%~dp0"

echo Building...
go build ./cmd/webapp
go build ./cmd/compute

popd

exit /b 0
