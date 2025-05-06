package devlog

import (
	"context"
	"strconv"
	"time"

	"connectrpc.com/connect"
)

type loggingHandlerStreamConn struct {
	connect.StreamingHandlerConn
	interceptor *loggingInterceptor
	ctx         context.Context
	start       time.Time
	reqID       string
	sentCount   int
	recvCount   int
}

func (i *loggingInterceptor) logStreamStart(ctx context.Context, lb *logBuilder, reqID string, conn connect.StreamingHandlerConn) {
	lb.writeMainHeader("STREAM START " + reqID)
	lb.writeKeyValue("Procedure", conn.Spec().Procedure)
	lb.writeKeyValue("Protocol", conn.Peer().Protocol)
	lb.writeKeyValue("Client", conn.Peer().Addr)
	lb.flush()

	lb.writeHeaders(conn.RequestHeader())
	lb.writeContextData(ctx, i.config.ContextExtractor)
}

func (i *loggingInterceptor) logStreamEnd(ctx context.Context, lb *logBuilder, conn *loggingHandlerStreamConn, err error) {
	lb.writeMainHeader("STREAM END " + conn.reqID)
	lb.writeKeyValue("Procedure", conn.Spec().Procedure)
	lb.writeKeyValue("Duration", time.Since(conn.start).String())
	lb.writeKeyValue("Sent", strconv.Itoa(conn.sentCount))
	lb.writeKeyValue("Received", strconv.Itoa(conn.recvCount))

	lb.writeContextData(ctx, i.config.ContextExtractor)

	if err != nil {
		lb.writeKeyValue("Error", err.Error())
		if connectErr, ok := err.(*connect.Error); ok {
			lb.writeKeyValue("ErrorCode", connectErr.Code().String())
		}
	}
}

func (c *loggingHandlerStreamConn) logStreamMessage(msg any, direction string, count int) {
	lb := getBuilder(c.interceptor.config)
	defer putBuilder(lb)

	lb.writeSubheader("STREAM " + direction + " #" + strconv.Itoa(count))
	if data, err := formatMessage(msg); err == nil {
		lb.writeRaw(string(data))
	} else {
		lb.writeRaw("[cannot format message: " + err.Error() + "]")
	}
	lb.writeTo(c.interceptor.logger.Writer())
}

func (c *loggingHandlerStreamConn) Send(msg any) error {
	err := c.StreamingHandlerConn.Send(msg)
	if err == nil {
		c.sentCount++
		c.logStreamMessage(msg, "SENT", c.sentCount)
	}
	return err
}

func (c *loggingHandlerStreamConn) Receive(msg any) error {
	err := c.StreamingHandlerConn.Receive(msg)
	if err == nil {
		c.recvCount++
		c.logStreamMessage(msg, "RECV", c.recvCount)
	}
	return err
}
