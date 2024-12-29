package gerror

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	CODE_OK              = 0
	CODE_GLOBAL_BOUNDARY = 10000
)

const (
	_MAGIC_PREFIX = "<cjey.xiao.gerror>"

	_FIELD_NAME_CODE    = "code"
	_FIELD_NAME_NAME    = "name"
	_FIELD_NAME_MESSAGE = "message"
)

// Encode encodes the code and msg and send to client
func Encode(code protoreflect.Enum, format string, a ...any) error {
	return New(code, format, a...).Encode()
}
func (e *GError) Encode() error {
	var prefix string
	if e.Message == "" {
		prefix = fmt.Sprintf("%s # %s[%d]", _MAGIC_PREFIX, e.Name, e.Code)
	} else {
		prefix = fmt.Sprintf("%s # %s[%d]: %s", _MAGIC_PREFIX, e.Name, e.Code, e.Message)
	}

	details, _ := structpb.NewValue(map[string]any{
		"code": e.Code, "name": e.Name, "message": e.Message,
	})
	sts, _ := status.New(codes.Unknown, prefix).WithDetails(details)
	return sts.Err()
}

// Decode decodes error returned by server to *GError
func Decode(err error) *GError {
	if err == nil {
		return &GError{Code: 0, Name: "OK"}
	}
	if e, ok := err.(*GError); ok {
		if e == nil {
			return &GError{Code: 0, Name: "OK"}
		}
		return e
	}

	sts := status.Convert(err)
	if sts.Code() == codes.OK {
		// OK always without message
		return &GError{Code: 0, Name: "OK"}
	}
	// try to extract server predefined error info
	if sts.Code() == codes.Unknown {
		if strings.HasPrefix(sts.Message(), _MAGIC_PREFIX) {
			// pattern hitted!
			if ds := sts.Details(); len(ds) == 1 {
				if val, ok := ds[0].(*structpb.Value); ok {
					if st := val.GetStructValue(); st != nil && st.Fields != nil {
						e := &GError{
							Code:    int32(st.Fields["code"].GetNumberValue()),
							Name:    st.Fields["name"].GetStringValue(),
							Message: st.Fields["message"].GetStringValue(),
						}
						if e.Code > 0 {
							return e
						}
					}
				}
			}
		}
	}
	// must be grpc error, convert code to negative
	return &GError{
		Code:    -1 * int32(sts.Code()),
		Name:    sts.Code().String(),
		Message: sts.Message(),
	}
}
