package httputil_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pthethanh/micro/status"
	"github.com/pthethanh/micro/util/httputil"
	"google.golang.org/grpc/codes"
)

func TestWrite(t *testing.T) {
	rc := httptest.NewRecorder()
	body := []byte(`{"name":"test"}`)
	rct := "application/json"
	httputil.Write(rc, rct, http.StatusOK, body)
	if bytes.Compare(rc.Body.Bytes(), body) != 0 {
		t.Errorf("got body: %s, want body: %s", rc.Body.Bytes(), body)
	}
	if ct := rc.Header().Get("Content-Type"); ct != rct {
		t.Errorf("got content type: %v, want content type: %v", ct, rct)
	}
	if status := rc.Result().StatusCode; status != http.StatusOK {
		t.Errorf("got status: %v, want status: %v", status, http.StatusOK)
	}
}

func TestWriteJSON(t *testing.T) {
	rc := httptest.NewRecorder()
	body := map[string]string{
		"name": "jack",
	}
	rct := "application/json"
	httputil.WriteJSON(rc, http.StatusOK, body)

	want := []byte(`{"name":"jack"}`)
	if bytes.Compare(rc.Body.Bytes(), want) != 0 {
		t.Errorf("got body: %s, want body: %s", rc.Body.Bytes(), want)
	}
	if ct := rc.Header().Get("Content-Type"); ct != rct {
		t.Errorf("got content type: %v, want content type: %v", ct, rct)
	}
	if status := rc.Result().StatusCode; status != http.StatusOK {
		t.Errorf("got status: %v, want status: %v", status, http.StatusOK)
	}

	// invalid - should write an internal error as json.
	rc = httptest.NewRecorder()
	// use chan to make json.Marshal fail.
	invalidData := make(chan int)
	httputil.WriteJSON(rc, http.StatusOK, invalidData)
	s, err := status.Parse(rc.Body.Bytes())
	if err != nil {
		t.Fatalf("got unexpected error: %v\n", err)
	}
	if !status.IsInternal(s.Err()) {
		t.Fatalf("got status=%v, want status is an internal error", s)
	}
}

func TestWriteError_Status(t *testing.T) {
	rc := httptest.NewRecorder()
	rct := "application/json"
	give := status.InvalidArgument("invalid request")
	httputil.WriteError(rc, http.StatusBadRequest, give)

	if ct := rc.Header().Get("Content-Type"); ct != rct {
		t.Errorf("got content type: %v, want content type: %v", ct, rct)
	}
	if status := rc.Result().StatusCode; status != http.StatusBadRequest {
		t.Errorf("got status: %v, want status: %v", status, http.StatusBadRequest)
	}
	got, err := status.Parse(rc.Body.Bytes())
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	if got.Err().Error() != give.Error() {
		t.Errorf("got err: %s, want err: %s", got.Err(), give)
	}
}

func TestWriteError_NormalError(t *testing.T) {
	rc := httptest.NewRecorder()
	rct := "application/json"
	give := errors.New("internal error")
	httputil.WriteError(rc, http.StatusInternalServerError, give)

	if ct := rc.Header().Get("Content-Type"); ct != rct {
		t.Errorf("got content type: %v, want content type: %v", ct, rct)
	}
	if status := rc.Result().StatusCode; status != http.StatusInternalServerError {
		t.Errorf("got status: %v, want status: %v", status, http.StatusInternalServerError)
	}
	wantMsg := fmt.Sprintf(`"message":"%s"`, give.Error())
	if !strings.Contains(rc.Body.String(), wantMsg) {
		t.Errorf("got err: %s, want err contains %s", rc.Body.String(), wantMsg)
	}
	wantCode := fmt.Sprintf(`"code":%d`, codes.Unknown)
	if !strings.Contains(rc.Body.String(), wantCode) {
		t.Errorf("got err: %s, want err contains %s", rc.Body.String(), wantCode)
	}
}
