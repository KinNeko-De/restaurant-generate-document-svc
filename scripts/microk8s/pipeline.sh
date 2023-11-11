cd .. && cd ..

go build -o ./bin/app ./cmd/document-generate-svc/main.go

sudo docker build -f ./build/dockerfile . -t restaurant-document-generate-svc:latest
sudo docker image ls
sudo docker tag restaurant-document-generate-svc localhost:32000/restaurant-document-generate-svc:latest
sudo docker push localhost:32000/restaurant-document-generate-svc:latest

kubectl apply --kustomize ./deployment/microk8s/overlays/dev

kubectl apply --kustomize ./deployment/microk8s/overlays/image
kubectl apply --kustomize ./deployment/microk8s/overlays/configerror
kubectl apply --kustomize ./deployment/microk8s/overlays/crash
kubectl apply --kustomize ./deployment/microk8s/overlays/notlive
kubectl apply --kustomize ./deployment/microk8s/overlays/notready


# kubectl logs -l app=document-generate-svc -tail 10 -n restaurant