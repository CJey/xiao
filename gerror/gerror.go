// Package gerror 把业务错误编码进grpc的带外错误，handler可直接返回*GError，调用方以Decode还原
package gerror

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	GrpcOK                 = codes.OK                 // 0
	GrpcCanceled           = codes.Canceled           // 1
	GrpcUnknown            = codes.Unknown            // 2
	GrpcInvalidArgument    = codes.InvalidArgument    // 3
	GrpcDeadlineExceeded   = codes.DeadlineExceeded   // 4
	GrpcNotFound           = codes.NotFound           // 5
	GrpcAlreadyExists      = codes.AlreadyExists      // 6
	GrpcPermissionDenied   = codes.PermissionDenied   // 7
	GrpcResourceExhausted  = codes.ResourceExhausted  // 8
	GrpcFailedPrecondition = codes.FailedPrecondition // 9
	GrpcAborted            = codes.Aborted            // 10
	GrpcOutOfRange         = codes.OutOfRange         // 11
	GrpcUnimplemented      = codes.Unimplemented      // 12
	GrpcInternal           = codes.Internal           // 13
	GrpcUnavailable        = codes.Unavailable        // 14
	GrpcDataLoss           = codes.DataLoss           // 15
	GrpcUnauthenticated    = codes.Unauthenticated    // 16
)

// GError 统一表示成功、业务错误与grpc错误，以Code的取值范围区分：
// 0为成功，(0, 10000)为局部业务错误，(10000, +∞)为全局业务错误，负数为取反的grpc状态码
type GError struct {
	Code    int32
	Name    string
	Message string
}

// New 以protobuf枚举创建业务错误，枚举值须为正且不得为CODE_GLOBAL_BOUNDARY，否则panic
func New(code protoreflect.Enum, format string, a ...any) *GError {
	var gerr = &GError{
		Code:    int32(code.Number()),
		Name:    code.(interface{ String() string }).String(),
		Message: fmt.Sprintf(format, a...),
	}
	if gerr.Code <= 0 {
		panic(fmt.Errorf("error code must be a positive enum number"))
	} else if gerr.Code == CODE_GLOBAL_BOUNDARY {
		panic(fmt.Errorf("reserved error code[%d] for local & global boundary", CODE_GLOBAL_BOUNDARY))
	}
	return gerr
}

func (e *GError) OK() bool {
	return e.Code == 0
}

// Equal 判断是否为给定的错误码，code须为codes.Code或protoreflect.Enum，其余类型一律返回false
func (e *GError) Equal(code any) bool {
	if c, ok := code.(codes.Code); ok {
		return e.Code == -1*int32(c) && e.Name == c.String()
	}
	if enum, ok := code.(protoreflect.Enum); ok {
		return e.Code == int32(enum.Number()) &&
			e.Name == enum.(interface{ String() string }).String()
	}
	return false
}

func (e *GError) IsLocal() bool {
	return e.Code > 0 && e.Code < int32(CODE_GLOBAL_BOUNDARY)
}

func (e *GError) IsGlobal() bool {
	return e.Code > int32(CODE_GLOBAL_BOUNDARY)
}

func (e *GError) IsGrpc() bool {
	return e.Code < 0
}

// IsGrpcTimeout 即grpc的DeadlineExceeded
func (e *GError) IsGrpcTimeout() bool {
	return e.Code == -1*int32(codes.DeadlineExceeded)
}

func (e *GError) IsGrpcUnknown() bool {
	return e.Code == -1*int32(codes.Unknown)
}

func (e *GError) IsGrpcUnavailable() bool {
	return e.Code == -1*int32(codes.Unavailable)
}

// Error 容许nil接收者：handler误返回类型化的nil时，grpc转换状态会调用Error
func (e *GError) Error() string {
	if e == nil {
		return "<nil>"
	}
	var str string
	if e.Code > 0 {
		str = fmt.Sprintf("%s[%d]", e.Name, e.Code)
	} else {
		str = e.Name
	}
	if e.Message == "" {
		return str
	} else {
		return str + ": " + e.Message
	}
}
