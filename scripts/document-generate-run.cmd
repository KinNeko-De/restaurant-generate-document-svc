:: starts the system under test
docker network create restaurant-dev-net

call document-generate-build.cmd

docker compose -f sut/docker-compose.yml up --build --remove-orphans --exit-code-from restaurant-document-svc

docker compose -f sut/docker-compose.yml down

docker image rm restaurant-document-svc
pause