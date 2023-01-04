module github.com/dvonthenen/rabbitmq-patterns

go 1.18

require (
	github.com/rabbitmq/amqp091-go v1.5.0
	k8s.io/klog/v2 v2.80.1
)

require github.com/go-logr/logr v1.2.0 // indirect

replace (
	github.com/gorilla/websocket => github.com/dvonthenen/websocket v1.5.1-0.20221123154619-09865dbf1be2
	github.com/koding/websocketproxy => github.com/dvonthenen/websocketproxy v0.0.0-20221223221552-50864d1ca0d3
	github.com/r3labs/sse/v2 => github.com/dvonthenen/sse/v2 v2.0.0-20221222171132-1daa5f8b774c
)

// replace github.com/dvonthenen/symbl-go-sdk => ../../dvonthenen/symbl-go-sdk
// replace github.com/dvonthenen/websocket => ../../dvonthenen/websocket
// replace github.com/dvonthenen/websocketproxy => ../../koding/websocketproxy
// replace github.com/dvonthenen/sse => ../../r3labs/sse
