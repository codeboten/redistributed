package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/codeboten/redistributed/propagation"
	"github.com/gorilla/mux"
	"github.com/opencensus-integrations/redigo/redis"
	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

var redisPool = &redis.Pool{
	Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", "redis:6379")
	},
	TestOnBorrow: func(c redis.Conn, t time.Time) error {
		if time.Since(t) < (5 * time.Minute) {
			return nil
		}
		_, err := c.Do("PING")
		return err
	},
}

func configureTracing() *jaeger.Exporter {
	agentEndpointURI := "jaeger:6831"
	collectorEndpointURI := "http://jaeger:14268/api/traces"

	je, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     agentEndpointURI,
		CollectorEndpoint: collectorEndpointURI,
		ServiceName:       "api",
	})
	if err != nil {
		log.Fatalf("Failed to create the Jaeger exporter: %v", err)
	}

	trace.RegisterExporter(je)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	return je
}

func reportsHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.FromContext(r.Context())
	defer span.End()
	conn := redisPool.GetWithContext(r.Context()).(redis.ConnWithContext)
	defer conn.CloseContext(r.Context())

	// Publish a message.
	publishCtx, publishSpan := trace.StartSpan(r.Context(), "publish")

	defer publishSpan.End()
	_, err := conn.DoContext(publishCtx, "PUBLISH", "mychannel1", propagation.Binary(publishSpan.SpanContext(), "message"))
	if err != nil {
		fmt.Printf("%s", err)
	}
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "reports generation request received : %v\n", vars["types"])
}

func main() {
	configureTracing()
	r := mux.NewRouter()
	r.HandleFunc("/reports", reportsHandler)
	h := &ochttp.Handler{Handler: r}

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":5000", h))
}
