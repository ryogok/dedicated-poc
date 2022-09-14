@echo off

pushd "%~dp0"

echo Pushing compute container
docker tag compute ryogokacr.azurecr.io/compute:v1.0
docker push ryogokacr.azurecr.io/compute:v1.0

echo.
echo Pushing webapp container
docker tag webapp ryogokacr.azurecr.io/webapp:v1.0
docker push ryogokacr.azurecr.io/webapp:v1.0

popd

exit /b 0
