package propagation

import (
	"bytes"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

func Binary(sc trace.SpanContext, message string) []byte {
	if sc == (trace.SpanContext{}) {
		return nil
	}
	output := [][]byte{propagation.Binary(sc), []byte(message)}

	return bytes.Join(output, []byte(""))
}

func FromBinary(bytes []byte) (trace.SpanContext, string, bool) {
	sc, ok := propagation.FromBinary(bytes[:29])
	if !ok {
		return sc, "", ok
	}
	return sc, string(bytes[29:]), ok
}
