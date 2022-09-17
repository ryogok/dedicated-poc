@echo off

pushd "%~dp0\.."

echo Deploying service.yaml...
kubectl apply -n ryogokpoc -f kubernetes\service.yaml

echo.
echo Updating containers...
kubectl rollout restart deployment webapp -n ryogokpoc
kubectl rollout restart deployment compute-p1 -n ryogokpoc
kubectl rollout restart deployment compute-p2 -n ryogokpoc

echo.
echo Done. Please check the deployment status with the following command:
echo kubectl get pods -n ryogokpoc

popd

exit /b 0
