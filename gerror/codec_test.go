package gerror

import (
	"context"
	"errors"
	"fmt"
	"net"
	"path"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
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

func TestDecodeWrapped(t *testing.T) {
	var encoded = Encode(typepb.Field_TYPE_STRING, "encoded")
	if g := Decode(fmt.Errorf("call foo: %w", encoded)); !g.Equal(typepb.Field_TYPE_STRING) {
		t.Errorf("wrapped status error decoded as %s", g)
	}
	if g := Decode(fmt.Errorf("call foo: %w", New(typepb.Field_TYPE_INT32, "raw"))); !g.Equal(typepb.Field_TYPE_INT32) {
		t.Errorf("wrapped *GError decoded as %s", g)
	}
}

func TestEncodeInvalidUTF8(t *testing.T) {
	var g = Decode(Encode(typepb.Field_TYPE_BYTES, "bad %s", "\xff\xfe"))
	if !g.Equal(typepb.Field_TYPE_BYTES) || !strings.Contains(g.Message, "\uFFFD") {
		t.Errorf("decoded as %s", g)
	}
}

// grpc错误经Decode后再次返回，须还原为原状态码
func TestGRPCStatusPassthrough(t *testing.T) {
	var g = Decode(status.Error(codes.Unavailable, "down"))
	if sts := status.Convert(g); sts.Code() != codes.Unavailable || sts.Message() != "down" {
		t.Errorf("passthrough as %s", sts)
	}
	if !Decode(g.Encode()).IsGrpcUnavailable() {
		t.Errorf("re-encoded as %s", Decode(g.Encode()))
	}
	if (&GError{Code: CODE_OK, Name: "OK"}).Encode() != nil {
		t.Error("OK should encode to nil")
	}
}

// 经真实的grpc往返，handler以各种形式返回错误
func TestOverGrpc(t *testing.T) {
	var errs = map[string]error{
		"direct":   New(typepb.Field_TYPE_STRING, "direct"),
		"wrapped":  fmt.Errorf("wrap: %w", New(typepb.Field_TYPE_INT32, "wrapped")),
		"encoded":  Encode(typepb.Field_TYPE_BOOL, "encoded"),
		"badutf8":  New(typepb.Field_TYPE_BYTES, "bad %s", "\xff\xfe"),
		"grpc":     &GError{Code: -int32(codes.Unavailable), Name: codes.Unavailable.String(), Message: "down"},
		"plain":    errors.New("plain"),
		"typednil": (*GError)(nil),
	}
	var checks = map[string]func(*GError) bool{
		"direct":   func(g *GError) bool { return g.Equal(typepb.Field_TYPE_STRING) },
		"wrapped":  func(g *GError) bool { return g.Equal(typepb.Field_TYPE_INT32) },
		"encoded":  func(g *GError) bool { return g.Equal(typepb.Field_TYPE_BOOL) },
		"badutf8":  func(g *GError) bool { return g.Equal(typepb.Field_TYPE_BYTES) },
		"grpc":     func(g *GError) bool { return g.IsGrpcUnavailable() && g.Message == "down" },
		"plain":    func(g *GError) bool { return g.IsGrpcUnknown() && g.Message == "plain" },
		"typednil": func(g *GError) bool { return g.IsGrpcUnknown() && g.Message == "<nil>" },
	}

	var lis = bufconn.Listen(1 << 20)
	var srv = grpc.NewServer(grpc.UnknownServiceHandler(func(_ any, ss grpc.ServerStream) error {
		var method, _ = grpc.MethodFromServerStream(ss)
		return errs[path.Base(method)]
	}))
	go srv.Serve(lis)
	defer srv.Stop()

	var conn, err = grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for name, check := range checks {
		var err = conn.Invoke(context.Background(), "/gerror.Test/"+name, &emptypb.Empty{}, &emptypb.Empty{})
		if g := Decode(err); !check(g) {
			t.Errorf("%s: decoded as %s (code %d)", name, g, g.Code)
		}
	}
}
