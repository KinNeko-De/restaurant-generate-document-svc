module github.com/kinneko-de/restaurant-document-generate-svc

go 1.20

require github.com/kinneko-de/protobuf-go v0.1.0

require (
	github.com/google/uuid v1.3.0
	github.com/kinneko-de/api-contract/golang/kinnekode/restaurant v0.2.5-document-request.10
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-rc.5
	github.com/kinneko-de/api-contract/golang/kinnekode/protobuf v0.2.5
	google.golang.org/grpc v1.57.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
)
