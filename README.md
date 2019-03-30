# redistributed
Library to support propagation of distributed tracing information for applications communicating over Redis channels.

## Requirements
* docker
* docker-compose

## Run example
```bash
git clone https://github.com/codeboten/redistributed.git
cd redistributed
docker-compose up

curl localhost:5000
open http://localhost:16686
```

## Usage
```go
import "github.com/codeboten/redistributed/propagation"

func publishMessage(w http.ResponseWriter, r *http.Request) {
	span := trace.FromContext(r.Context())
	conn := redisPool.GetWithContext(r.Context()).(redis.ConnWithContext)
	publishCtx, publishSpan := trace.StartSpan(r.Context(), "publish")

	_, err := conn.DoContext(publishCtx, "PUBLISH", "mychannel1", redispropagation.Binary(publishSpan.SpanContext(), "message"))
}

func receiveMessage(ctx context.Context) {
	pubsub := client.Subscribe("mychannel1")
	ch := pubsub.Channel()
	for msg := range ch {
		spanContext, message, ok := redispropagation.FromBinary(payload)
		if ok {
			_, span := trace.StartSpanWithRemoteParent(ctx, "generateReport", spanContext)
		}
		fmt.Printf("Message received:%s", message)
	}
}
```