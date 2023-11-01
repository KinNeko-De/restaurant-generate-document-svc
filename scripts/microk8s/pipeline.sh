cd .. && cd ..

go build -o ./bin/app ./cmd/document-generate-svc/main.go

sudo docker build -f ./build/dockerfile . -t restaurant-document-generate-svc:latest
sudo docker image ls
sudo docker tag restaurant-document-generate-svc localhost:32000/restaurant-document-generate-svc:latest
sudo docker push localhost:32000/restaurant-document-generate-svc:latest

kubectl apply --kustomize ./deployment/microk8s/overlays/dev

# kubectl logs -l app=document-generate-svc -n restaurant