package main

import (
	"context"
	"fmt"
	"log"

	"github.com/codeboten/redistributed/propagation"
	"github.com/go-redis/redis"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

func configureTracing() *jaeger.Exporter {
	// Port details: https://www.jaegertracing.io/docs/getting-started/
	agentEndpointURI := "jaeger:6831"
	collectorEndpointURI := "http://jaeger:14268/api/traces"

	je, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     agentEndpointURI,
		CollectorEndpoint: collectorEndpointURI,
		ServiceName:       "report-generator",
	})
	if err != nil {
		log.Fatalf("Failed to create the Jaeger exporter: %v", err)
	}

	// And now finally register it as a Trace Exporter
	trace.RegisterExporter(je)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	return je
}

func generateReport(ctx context.Context, payload []byte) {
	spanContext, message, ok := propagation.FromBinary(payload)
	if ok {
		_, span := trace.StartSpanWithRemoteParent(ctx, "generateReport", spanContext)
		defer span.End()
	} else {
		_, span := trace.StartSpan(ctx, "generateReport")
		message = string(payload)
		defer span.End()
	}
	fmt.Printf("Message received:%s", message)
}

func startListener(ctx context.Context) {
	_, span := trace.StartSpan(ctx, "GenerateMonthlyReport")
	defer span.End()
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pubsub := client.Subscribe("mychannel1")

	// Wait for confirmation that subscription is created before publishing anything.
	// _, err := pubsub.Receive()
	// if err != nil {
	// 	fmt.Printf("%s", err)
	// }

	// Go channel which receives messages.
	ch := pubsub.Channel()

	// time.AfterFunc(100*time.Second, func() {
	// 	// When pubsub is closed channel is closed too.
	// 	_ = pubsub.Close()
	// })

	// Consume messages.
	for msg := range ch {
		generateReport(ctx, []byte(msg.Payload))
	}
}

func main() {
	exporter := configureTracing()
	ctx, span := trace.StartSpan(context.Background(), "something")

	startListener(ctx)
	span.End()
	exporter.Flush()
}
