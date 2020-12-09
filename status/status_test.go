package status_test

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/pthethanh/micro/status"
	"google.golang.org/grpc/codes"
)

func TestErrors(t *testing.T) {
	if !status.IsOK(status.OK("").Err()) {
		t.Errorf("got !OK, want OK")
	}
	if !status.IsCanceled(status.Canceled("")) {
		t.Errorf("got !Canceled, want Canceled")
	}
	if !status.IsUnknown(status.Unknown("")) {
		t.Errorf("got !Unknown, want Unknown")
	}
	if !status.IsInvalidArgument(status.InvalidArgument("")) {
		t.Errorf("got !InvalidArgument, want InvalidArgument")
	}
	if !status.IsDeadlineExceeded(status.DeadlineExceeded("")) {
		t.Errorf("got !DeadlineExceeded, want DeadlineExceeded")
	}
	if !status.IsNotFound(status.NotFound("")) {
		t.Errorf("got !NotFound, want NotFound")
	}
	if !status.IsAlreadyExists(status.AlreadyExists("")) {
		t.Errorf("got !AlreadyExists, want AlreadyExists")
	}
	if !status.IsPermissionDenied(status.PermissionDenied("")) {
		t.Errorf("got !PermissionDenied, want PermissionDenied")
	}
	if !status.IsResourceExhausted(status.ResourceExhausted("")) {
		t.Errorf("got !ResourceExhausted, want ResourceExhausted")
	}
	if !status.IsFailedPrecondition(status.FailedPrecondition("")) {
		t.Errorf("got !FailedPrecondition, want FailedPrecondition")
	}
	if !status.IsAborted(status.Aborted("")) {
		t.Errorf("got !Aborted, want Aborted")
	}
	if !status.IsOutOfRange(status.OutOfRange("")) {
		t.Errorf("got !OutOfRange, want OutOfRange")
	}
	if !status.IsUnimplemented(status.Unimplemented("")) {
		t.Errorf("got !Unimplemented, want Unimplemented")
	}
	if !status.IsInternal(status.Internal("")) {
		t.Errorf("got !Internal, want Internal")
	}
	if !status.IsUnavailable(status.Unavailable("")) {
		t.Errorf("got !Unavailable, want Unavailable")
	}
	if !status.IsDataLoss(status.DataLoss("")) {
		t.Errorf("got !DataLoss, want DataLoss")
	}
	if !status.IsUnauthenticated(status.Unauthenticated("")) {
		t.Errorf("got !Unauthenticated, want Unauthenticated")
	}
	if !status.IsUnavailable(status.Unavailable("")) {
		t.Errorf("got !Unavailable, want Unavailable")
	}
	if !status.IsUnavailable(status.Unavailable("")) {
		t.Errorf("got !Unavailable, want Unavailable")
	}
	if !status.IsUnavailable(status.New(codes.Unavailable, "").Err()) {
		t.Errorf("new - got !Unavailable, want Unavailable")
	}
	if !status.IsUnavailable(status.New(codes.Unavailable, "server %d unavailable", 1).Err()) {
		t.Errorf("new with fmt - got !Unavailable, want Unavailable")
	}
}

func TestHTTPStatusCodeCovert(t *testing.T) {

	cases := []struct {
		give codes.Code
		want int
	}{
		{
			give: codes.OK,
			want: http.StatusOK,
		},
		{
			give: codes.Canceled,
			want: http.StatusRequestTimeout,
		},
		{
			give: codes.Unknown,
			want: http.StatusInternalServerError,
		},
		{
			give: codes.InvalidArgument,
			want: http.StatusBadRequest,
		},
		{
			give: codes.DeadlineExceeded,
			want: http.StatusGatewayTimeout,
		},
		{
			give: codes.NotFound,
			want: http.StatusNotFound,
		},
		{
			give: codes.AlreadyExists,
			want: http.StatusConflict,
		},
		{
			give: codes.PermissionDenied,
			want: http.StatusForbidden,
		},
		{
			give: codes.Unauthenticated,
			want: http.StatusUnauthorized,
		},
		{
			give: codes.ResourceExhausted,
			want: http.StatusTooManyRequests,
		},
		{
			give: codes.FailedPrecondition,
			want: http.StatusBadRequest,
		},
		{
			give: codes.Aborted,
			want: http.StatusConflict,
		},
		{
			give: codes.OutOfRange,
			want: http.StatusBadRequest,
		},
		{
			give: codes.Unimplemented,
			want: http.StatusNotImplemented,
		},
		{
			give: codes.Internal,
			want: http.StatusInternalServerError,
		},
		{
			give: codes.Unavailable,
			want: http.StatusServiceUnavailable,
		},
		{
			give: codes.DataLoss,
			want: http.StatusInternalServerError,
		},
		{
			give: codes.Code(9999),
			want: http.StatusInternalServerError,
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s", c.give.String()), func(t *testing.T) {
			if s := status.HTTPStatusFromCode(c.give); s != c.want {
				t.Errorf("got status=%d, want status=%d", s, c.want)
			}
		})
	}
}

func TestConvert(t *testing.T) {
	s := status.NotFound("name not found, value: %v", "test")
	if code := status.Convert(s).Code(); code != codes.NotFound {
		t.Errorf("got code=%d, want code=%d", code, codes.NotFound)
	}
	if code := status.Convert(status.Error(codes.AlreadyExists, "data already exist")).Code(); code != codes.AlreadyExists {
		t.Errorf("got code=%d, want code=%d", code, codes.AlreadyExists)
	}
	if code := status.Convert(errors.New("some error")).Code(); code != codes.Unknown {
		t.Errorf("got code=%d, want code=%d", code, codes.Unknown)
	}
	sts, ok := status.FromError(errors.New("some error"))
	if ok {
		t.Errorf("got convert ok , want failed to convert")
	}
	if sts.Code() != codes.Unknown {
		t.Errorf("got code=%d , want code=%d", sts.Code(), codes.Unknown)
	}

}

func TestConvertJSON(t *testing.T) {
	b := status.JSON(status.NotFound("data not found, name=%s", "test"))
	exp := fmt.Sprintf(`"code":%d`, codes.NotFound)
	if !strings.Contains(string(b), exp) {
		t.Errorf("got data=%s, want data contains %s", b, exp)
	}
	exp = fmt.Sprintf(`"message":"data not found, name=%s"`, "test")
	if !strings.Contains(string(b), exp) {
		t.Errorf("got data=%s, want data contains %v", b, exp)
	}
	// custom error
	b = status.JSON(status.New(status.Code(999), "data not found, name=%s", "999").Err())
	exp = fmt.Sprintf(`"code":%d`, 999)
	if !strings.Contains(string(b), exp) {
		t.Errorf("got data=%s, want data contains %s", b, exp)
	}
	exp = fmt.Sprintf(`"message":"data not found, name=%s"`, "999")
	if !strings.Contains(string(b), exp) {
		t.Errorf("got data=%s, want data contains %v", b, exp)
	}
}
