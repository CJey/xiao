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

// Encode 以New创建业务错误并编码为grpc错误
func Encode(code protoreflect.Enum, format string, a ...any) error {
	return New(code, format, a...).Encode()
}

// Encode 编码为grpc错误，成功时返回nil
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

// Decode 将grpc调用返回的错误还原为*GError：nil为成功，非本包编码的错误按grpc状态码还原
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
		// 成功时不带消息
		return &GError{Code: CODE_OK, Name: "OK"}
	}
	// 包装过的错误经status.FromError后message会被换成包装后的全文，业务错误仅以details为准
	if sts.Code() == codes.Unknown {
		if e := fromDetails(sts.Details()); e != nil {
			return e
		}
	}
	// 其余均为grpc错误，状态码取反存放
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
