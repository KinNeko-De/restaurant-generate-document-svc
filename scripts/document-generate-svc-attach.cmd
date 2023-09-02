:: attach to the system under test, only works if you define a CMD instead of an ENTRYPOINT
docker network create restaurant

call document-generate-svc-build.cmd

docker compose -f sut/docker-compose.yml build

docker run -v %cd%\sut\log/:/app/log/ -v %cd%\sut\run/:/app/run/ -v %cd%\sut\template/:/app/template/ -t -i --name restaurant-document-generate-svc restaurant-document-generate-svc bash

docker rm restaurant-document-generate-svc

docker image rm restaurant-document-generate-svc

pause