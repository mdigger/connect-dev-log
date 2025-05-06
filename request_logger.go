package devlog

import (
	"context"
	"reflect"
	"time"

	"connectrpc.com/connect"
)

func (i *loggingInterceptor) logRequest(ctx context.Context, lb *logBuilder, reqID string, req connect.AnyRequest) {
	lb.writeMainHeader("REQUEST " + reqID)
	lb.writeKeyValue("Procedure", req.Spec().Procedure)
	lb.writeKeyValue("Protocol", req.Peer().Protocol)
	lb.writeKeyValue("Client", req.Peer().Addr)
	lb.flush()

	lb.writeContextData(ctx, i.config.ContextExtractor)
	lb.writeHeaders(req.Header())

	if req.Any() != nil {
		lb.writeSubheader("Request Body")
		if data, err := formatMessage(req.Any()); err == nil {
			lb.buf.Write(data)
			lb.buf.WriteByte('\n')
		} else {
			lb.buf.WriteString("[cannot format request: " + err.Error() + "]\n")
		}
	}
}

func (i *loggingInterceptor) logResponse(ctx context.Context, lb *logBuilder, reqID string, req connect.AnyRequest, resp connect.AnyResponse, err error, duration time.Duration) {
	lb.writeMainHeader("RESPONSE " + reqID)
	lb.writeKeyValue("Procedure", req.Spec().Procedure)
	lb.writeKeyValue("Duration", duration.String())

	if err != nil {
		if connectErr, ok := err.(*connect.Error); ok {
			lb.writeKeyValue("Error", connectErr.Message())
			lb.writeKeyValue("Code", connectErr.Code().String())
		} else {
			lb.writeKeyValue("Error", err.Error())
		}
	}
	lb.flush()

	lb.writeContextData(ctx, i.config.ContextExtractor)

	if reflect.ValueOf(resp).IsNil() {
		return
	}

	lb.writeHeaders(resp.Header())

	if resp.Any() != nil {
		lb.writeSubheader("Response Body")
		if data, err := formatMessage(resp.Any()); err == nil {
			lb.buf.Write(data)
			lb.buf.WriteByte('\n')
		} else {
			lb.buf.WriteString("[cannot format response: " + err.Error() + "]\n")
		}
	}
}
