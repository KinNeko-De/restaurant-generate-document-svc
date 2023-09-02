:: starts the system under test
docker network create restaurant

call document-generate-svc-build.cmd

docker compose -f sut/docker-compose.yml up --build --remove-orphans --exit-code-from restaurant-document-generate-svc

docker compose -f sut/docker-compose.yml down

docker image rm restaurant-document-generate-svc
pause