package protocol

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const maxLineSize = 1 << 20 // 1 MiB

// readLine reads the next non-blank line from r.
// Returns an error if the line exceeds maxLineSize.
func readLine(r *bufio.Reader) ([]byte, error) {
	for {
		line, err := r.ReadBytes('\n')
		if err != nil && len(line) == 0 {
			return nil, err
		}
		if len(line) > maxLineSize {
			return nil, fmt.Errorf("line exceeds max size (%d bytes)", maxLineSize)
		}
		line = bytes.TrimRight(line, "\r\n")
		if len(line) == 0 {
			if err != nil {
				return nil, err
			}
			continue // skip blank lines
		}
		return line, nil
	}
}

// Encode marshals msg as JSON and writes it to w followed by a newline.
func Encode(w io.Writer, msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

// Decode reads one JSON line from r and returns it as a map.
func Decode(r *bufio.Reader) (map[string]any, error) {
	line, err := readLine(r)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// DecodeInto reads one JSON line from r and unmarshals it into target.
func DecodeInto(r *bufio.Reader, target any) error {
	line, err := readLine(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(line, target)
}

// decodeAs is a generic helper: unmarshal line into T and return the value.
func decodeAs[T any](line []byte) (any, error) {
	var v T
	if err := json.Unmarshal(line, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// DecodeTyped reads one JSON line from r, inspects the "type" field,
// and returns the correct concrete struct value.
func DecodeTyped(r *bufio.Reader) (any, error) {
	line, err := readLine(r)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &envelope); err != nil {
		return nil, err
	}

	switch envelope.Type {
	case TypeReady:
		return decodeAs[ReadyEvent](line)
	case TypeMessage, TypeEdit:
		return decodeAs[MessageEvent](line)
	case TypeReaction:
		return decodeAs[ReactionEvent](line)
	case TypeResponse:
		return decodeAs[ResponseEvent](line)
	case TypeError:
		return decodeAs[ErrorEvent](line)
	case TypeShutdownComplete:
		return decodeAs[ShutdownCompleteEvent](line)
	case TypePing:
		return decodeAs[PingEvent](line)
	case TypePong:
		return decodeAs[PongEvent](line)
	case TypeStart:
		return decodeAs[StartCommand](line)
	case TypeSend:
		return decodeAs[SendCommand](line)
	case TypeSendMedia:
		return decodeAs[SendMediaCommand](line)
	case TypeReact:
		return decodeAs[ReactCommand](line)
	case TypeTyping:
		return decodeAs[TypingCommand](line)
	case TypeShutdown:
		return decodeAs[ShutdownCommand](line)
	default:
		return nil, fmt.Errorf("unknown protocol type: %q", envelope.Type)
	}
}
