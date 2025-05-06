package devlog

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func formatMessage(msg any) ([]byte, error) {
	if msg == nil {
		return []byte("(none)"), nil
	}

	switch v := msg.(type) {
	case proto.Message:
		return prototext.MarshalOptions{
			Multiline: true,
			Indent:    "  ",
		}.Marshal(v)
	default:
		return json.MarshalIndent(v, "", "  ")
	}
}
