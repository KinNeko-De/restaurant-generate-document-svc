:: starts the system under test
docker network create restaurant-dev-net

call generate-document-build.cmd

docker compose -f sut/docker-compose.yml up --build --remove-orphans --exit-code-from restaurant-generate-document-svc

docker compose -f sut/docker-compose.yml down

docker image rm restaurant-generate-document-svc
pause