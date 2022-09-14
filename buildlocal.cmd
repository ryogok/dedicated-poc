@echo off

pushd "%~dp0"

echo Building...
go build webapp.go
go build compute.go

popd

exit /b 0
