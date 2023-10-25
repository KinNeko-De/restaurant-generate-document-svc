module github.com/kinneko-de/restaurant-document-generate-svc

go 1.21

toolchain go1.21.1

require github.com/kinneko-de/protobuf-go v0.2.0

require (
	github.com/google/uuid v1.3.1
	github.com/kinneko-de/api-contract/golang/kinnekode/restaurant v0.0.1
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.1
	github.com/kinneko-de/api-contract/golang/kinnekode/protobuf v0.2.6
	go.opentelemetry.io/otel v1.19.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.42.0
	go.opentelemetry.io/otel/sdk v1.19.0
	google.golang.org/grpc v1.58.3
)

require (
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.1 // indirect
	go.opentelemetry.io/otel/trace v1.19.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230920204549-e6e6cdab5c13 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/go-logr/zerologr v1.2.3
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/rs/zerolog v1.31.0
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.42.0
	go.opentelemetry.io/otel/metric v1.19.0
	go.opentelemetry.io/otel/sdk/metric v1.19.0
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230920204549-e6e6cdab5c13 // indirect
)
