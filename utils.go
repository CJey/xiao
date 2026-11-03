package xiao

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"

	"github.com/goccy/go-json"
)

// RawJson 刻意不实现fmt.Stringer：go1.27起json.RawMessage即jsontext.Value，带String方法，
// 经zap.Any输出时会被当作字符串转义
type RawJson []byte

func (r RawJson) MarshalJSON() ([]byte, error) {
	if r == nil {
		return []byte("null"), nil
	}
	return r, nil
}

func (r *RawJson) UnmarshalJSON(data []byte) error {
	if r == nil {
		return errors.New("xiao.RawJson: UnmarshalJSON on nil pointer")
	}
	*r = append((*r)[0:0], data...)
	return nil
}

// Json 将msg序列化为可原样嵌入日志的RawJson；[]byte视为已编码的json，序列化失败时为null
func Json(msg any) RawJson {
	if raw, yes := msg.([]byte); yes {
		return RawJson(raw)
	}
	js, _ := json.Marshal(msg)
	return RawJson(js)
}

// JsonPretty 返回缩进后的json，[]byte与string视为已编码的json，无法解析时原样返回；prefix[0]为首行之外各行的前缀
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
