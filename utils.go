package xiao

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/goccy/go-json"
)

func Json(msg any) json.RawMessage {
	if raw, yes := msg.([]byte); yes {
		return json.RawMessage(raw)
	}
	js, _ := json.Marshal(msg)
	return json.RawMessage(js)
}

func JsonPretty(msg any, prefix ...string) string {
	pre := ""
	if len(prefix) > 0 {
		pre = prefix[0]
	}
	if raw, yes := msg.([]byte); yes {
		var buf bytes.Buffer
		if err := json.Indent(&buf, raw, pre, "  "); err != nil {
			return string(raw)
		} else {
			return buf.String()
		}
	}
	if raw, yes := msg.(string); yes {
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(raw), pre, "  "); err != nil {
			return raw
		} else {
			return buf.String()
		}
	}
	js, _ := json.MarshalIndent(msg, pre, "  ")
	return string(js)
}

func Gzip(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	var zw = gzip.NewWriter(&buf)
	if _, err := zw.Write(input); err != nil {
		return nil, err
	} else if err = zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Gunzip(input []byte) ([]byte, error) {
	var zr, err = gzip.NewReader(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, zr); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
