package gerror

import (
	"testing"

	"google.golang.org/protobuf/types/known/typepb"
)

func TestEncode1(t *testing.T) {
	err := Encode(typepb.Field_TYPE_STRING, "use any pb enum for test")
	t.Logf("origin err = %s", err)
	gerr := Decode(err)
	if !gerr.OK() && gerr.Equal(typepb.Field_TYPE_STRING) {
		t.Logf("gerror = %s", gerr)
	} else {
		t.Errorf("bad codec")
	}
}

func TestEncode2(t *testing.T) {
	gerr := Decode(nil)
	if gerr.OK() {
		t.Logf("%s", gerr)
	} else {
		t.Errorf("bad codec")
	}
}
