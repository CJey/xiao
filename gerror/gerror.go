// gerror提供用于编解码grpc错误的对象和方法，用于将业务错误编码在带外错误中，
// 从而让grpc的使用方式上更加符合go style
package gerror

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// 枚举grpc预定义的code
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

// composite all error categories, with no conflict encoding
// include: ok, local error, global error, and grpc error
type GError struct {
	// 0: ok
	// (0, 10000): local error
	// (10000, +inf): global error
	// (-inf, 0): external error
	Code    int32
	Name    string
	Message string
}

func New(code protoreflect.Enum, format string, a ...any) *GError {
	var gerr = &GError{
		Code:    int32(code.Number()),
		Name:    code.(interface{ String() string }).String(),
		Message: fmt.Sprintf(format, a...),
	}
	if gerr.Code <= 0 {
		panic(fmt.Errorf("error code must be a positive enum number"))
	} else if gerr.Code == CODE_GLOBAL_BOUNDARY {
		panic(fmt.Errorf("reserverd error code[%d] for local & global boundary", CODE_GLOBAL_BOUNDARY))
	}
	return gerr
}

// OK means ok
func (e *GError) OK() bool {
	return e.Code == 0
}

// Equal used to compare the given code, only support grpc.codes.Code or protoreflect.Enum
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

// is local error
func (e *GError) IsLocal() bool {
	return e.Code > 0 && e.Code < int32(CODE_GLOBAL_BOUNDARY)
}

// is global error
func (e *GError) IsGlobal() bool {
	return e.Code > int32(CODE_GLOBAL_BOUNDARY)
}

// is grpc error
func (e *GError) IsGrpc() bool {
	return e.Code < 0
}

// is grpc timeout error
func (e *GError) IsGrpcTimeout() bool {
	return e.Code == -1*int32(codes.DeadlineExceeded)
}

// is grpc unknown error
func (e *GError) IsGrpcUnknown() bool {
	return e.Code == -1*int32(codes.Unknown)
}

// Error implement error interface
func (e *GError) Error() string {
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
