:: attach to the system under test, only works if you define a CMD instead of an ENTRYPOINT
docker network create restaurant-dev-net

call build-main.cmd

docker compose -f sut/docker-compose.yml build

docker run -v %cd%\sut\log/:/app/log/ -t -i --name restaurant-document-svc restaurant-document-svc bash

docker rm restaurant-document-svc

docker image rm restaurant-document-svc

pause