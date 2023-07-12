:: attach to the system under test, only works if you define a CMD instead of an ENTRYPOINT
docker network create restaurant-dev-net

call sut-build.cmd

docker compose -f sut/docker-compose.yml build

docker run -v %cd%\sut\log/:/app/log/ -v %cd%\sut\run/:/app/run/ -v %cd%\sut\template/:/app/template/ -t -i --name restaurant-generate-document-svc restaurant-generate-document-svc bash

docker rm restaurant-generate-document-svc

docker image rm restaurant-generate-document-svc

pause