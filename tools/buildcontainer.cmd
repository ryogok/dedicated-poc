@echo off

pushd "%~dp0\.."

echo Building compute container
docker build -t compute -f ./cmd/compute/Dockerfile .

echo.
echo Building webapp container
docker build -t webapp -f ./cmd/webapp/Dockerfile .

popd

exit /b 0
