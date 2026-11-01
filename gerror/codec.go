package gerror

import (
	"errors"
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
	return e.GRPCStatus().Err()
}

// GRPCStatus 供grpc转换错误时调用，故handler可直接返回*GError。
// 业务错误编码进带外的details，grpc错误还原为原状态码
func (e *GError) GRPCStatus() *status.Status {
	if e == nil || e.Code == CODE_OK {
		return nil
	}
	if e.Code < 0 {
		return status.New(codes.Code(-e.Code), e.Message)
	}

	// details随status整体序列化，含非法UTF-8时会被grpc整体丢弃
	var name = strings.ToValidUTF8(e.Name, "\uFFFD")
	var msg = strings.ToValidUTF8(e.Message, "\uFFFD")
	var prefix = fmt.Sprintf("%s # %s[%d]", _MAGIC_PREFIX, name, e.Code)
	if msg != "" {
		prefix += ": " + msg
	}

	var sts = status.New(codes.Unknown, prefix)
	var details, err = structpb.NewValue(map[string]any{
		_FIELD_NAME_CODE: e.Code, _FIELD_NAME_NAME: name, _FIELD_NAME_MESSAGE: msg,
	})
	if err == nil {
		if withDetails, err := sts.WithDetails(details); err == nil {
			sts = withDetails
		}
	}
	return sts
}

// Decode decodes error returned by server to *GError
func Decode(err error) *GError {
	if err == nil {
		return &GError{Code: CODE_OK, Name: "OK"}
	}
	if gerr, ok := errors.AsType[*GError](err); ok {
		if gerr == nil {
			return &GError{Code: CODE_OK, Name: "OK"}
		}
		return gerr
	}

	sts := status.Convert(err)
	if sts.Code() == codes.OK {
		// OK always without message
		return &GError{Code: CODE_OK, Name: "OK"}
	}
	// 包装过的错误经status.FromError后message会被换成包装后的全文，业务错误只认details
	if sts.Code() == codes.Unknown {
		if e := fromDetails(sts.Details()); e != nil {
			return e
		}
	}
	// must be grpc error, convert code to negative
	return &GError{
		Code:    -1 * int32(sts.Code()),
		Name:    sts.Code().String(),
		Message: sts.Message(),
	}
}

func fromDetails(ds []any) *GError {
	if len(ds) != 1 {
		return nil
	}
	var val, ok = ds[0].(*structpb.Value)
	if !ok {
		return nil
	}
	var st = val.GetStructValue()
	if st == nil {
		return nil
	}
	var e = &GError{
		Code:    int32(st.Fields[_FIELD_NAME_CODE].GetNumberValue()),
		Name:    st.Fields[_FIELD_NAME_NAME].GetStringValue(),
		Message: st.Fields[_FIELD_NAME_MESSAGE].GetStringValue(),
	}
	if e.Code <= 0 || e.Name == "" {
		return nil
	}
	return e
}
