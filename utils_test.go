package xiao

import (
	stdjson "encoding/json"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

// 日志中须原样嵌入json，不得被转义为字符串
func TestRawJsonInLog(t *testing.T) {
	var buf = captureLogger(t)
	NamedContext("json").Info("msg", "data", Json(map[string]int{"a": 1}))
	if !strings.Contains(buf.String(), `"data":{"a":1}`) {
		t.Errorf("RawJson is not embedded as json: %q", buf)
	}
}

func TestRawJsonCodec(t *testing.T) {
	type pair struct {
		A RawJson `json:"a"`
		B RawJson `json:"b"`
	}
	for name, marshal := range map[string]func(any) ([]byte, error){
		"goccy": json.Marshal, "std": stdjson.Marshal,
	} {
		var js, err = marshal(pair{A: Json([]int{1, 2})})
		if err != nil || string(js) != `{"a":[1,2],"b":null}` {
			t.Errorf("%s marshal: %s, %v", name, js, err)
		}
	}

	var got pair
	if err := json.Unmarshal([]byte(`{"a":[1,2],"b":{"x":"y"}}`), &got); err != nil ||
		string(got.A) != `[1,2]` || string(got.B) != `{"x":"y"}` {
		t.Errorf("unmarshal: %+v, %v", got, err)
	}
}
