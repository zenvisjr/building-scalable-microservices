module github.com/zenvisjr/building-scalable-microservices

go 1.23.3

// Local module replacements
replace github.com/zenvisjr/building-scalable-microservices/logger => ./logger

replace github.com/zenvisjr/building-scalable-microservices/account => ./account

replace github.com/zenvisjr/building-scalable-microservices/catalog => ./catalog

replace github.com/zenvisjr/building-scalable-microservices/order => ./order

replace github.com/zenvisjr/building-scalable-microservices/mail => ./mail

require (
	github.com/99designs/gqlgen v0.17.77
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.10.9
	github.com/nats-io/nats.go v1.43.0
	github.com/olivere/elastic/v7 v7.0.32
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.17.0
	github.com/segmentio/ksuid v1.0.4
	github.com/sendgrid/sendgrid-go v3.16.1+incompatible
	github.com/vektah/gqlparser/v2 v2.5.30
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.74.2
	google.golang.org/protobuf v1.36.6

)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
)
