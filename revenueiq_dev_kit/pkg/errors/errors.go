package errors

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapGRPCErrToHTTPStatus inspects a gRPC error and returns the corresponding HTTP status code
// along with the error description message.
func MapGRPCErrToHTTPStatus(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, err.Error()
	}

	msg := st.Message()
	switch st.Code() {
	case codes.InvalidArgument:
		return http.StatusBadRequest, msg
	case codes.Unauthenticated:
		return http.StatusUnauthorized, msg
	case codes.PermissionDenied:
		return http.StatusForbidden, msg
	case codes.NotFound:
		return http.StatusNotFound, msg
	case codes.AlreadyExists:
		return http.StatusConflict, msg
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests, msg
	case codes.Unimplemented:
		return http.StatusNotImplemented, msg
	case codes.Unavailable:
		return http.StatusServiceUnavailable, msg
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, msg
	default:
		return http.StatusInternalServerError, msg
	}
}
