package httputil_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pthethanh/micro/util/httputil"
)

func TestMock(t *testing.T) {
	handlers := httputil.MustReadMockFromFile("testdata/mock.json")
	srv := httptest.NewServer(httputil.Mock(handlers...))
	defer srv.Close()

	cases := []struct {
		name    string
		headers map[string]string
		method  string
		path    string

		code int
		// too lazy, let assume it's a map for easier to test.
		body       map[string]interface{}
		resHeaders map[string]string
		verify     func(body map[string]interface{}, t *testing.T)
	}{
		{
			name:   "get users ok - data from file",
			method: http.MethodGet,
			path:   "/users",
			code:   http.StatusOK,
			body: map[string]interface{}{
				"users": []map[string]interface{}{
					{
						"id":   "1",
						"name": "jack",
						"age":  22,
					},
					{
						"id":   "2",
						"name": "mia",
						"age":  16,
					},
				},
			},
			resHeaders: map[string]string{"session_id": "123"},
			verify: func(body map[string]interface{}, t *testing.T) {
				if v, ok := body["users"].([]interface{}); !ok || len(v) != 2 {
					t.Errorf("got len(users)=%v, want len(users)=%v", len(v), 2)
				}
			},
		},
		{
			name:   "delete users not found",
			method: http.MethodDelete,
			path:   "/users",
			code:   http.StatusNotFound,
			body: map[string]interface{}{
				"code": 5,
			},
		},
		{
			name:   "get user detail ok",
			method: http.MethodGet,
			path:   "/users/1",
			code:   http.StatusOK,
			body: map[string]interface{}{
				"id":   "1",
				"name": "jack",
				"age":  22,
			},
		},
		{
			name:    "delete users unauthorized - with header",
			method:  http.MethodDelete,
			path:    "/users/1",
			code:    http.StatusUnauthorized,
			headers: map[string]string{"authorization": "not_ok"},
			body: map[string]interface{}{
				"code": 16,
			},
		},
		{
			name:       "delete users ok with response header",
			method:     http.MethodDelete,
			path:       "/users/1",
			code:       http.StatusOK,
			headers:    map[string]string{"authorization": "ok"},
			resHeaders: map[string]string{"session_id": "123"},
			body: map[string]interface{}{
				"code": 0,
			},
		},
		{
			name:   "get from file with response code",
			method: http.MethodGet,
			path:   "/employees",
			code:   http.StatusNotFound,
			body: map[string]interface{}{
				"code":    5,
				"message": "not found",
			},
		},
		{
			name:   "get from file - not found the file ",
			method: http.MethodGet,
			path:   "/employees/1",
			code:   http.StatusNotFound,
			body: map[string]interface{}{
				"code": 2,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(c.method, srv.URL+c.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			for k, v := range c.headers {
				req.Header.Set(k, v)
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
			}
			defer res.Body.Close()
			if res.StatusCode != c.code {
				t.Fatalf("got status=%d, want status=%d", res.StatusCode, c.code)
			}
			var body map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if c.verify == nil {
				for k, v := range c.body {
					if gv, ok := body[k]; !ok || fmt.Sprintf("%v", v) != fmt.Sprintf("%v", gv) {
						t.Fatalf("got body[%s]=%v, want body[%s]=%v", k, gv, k, v)
					}
				}
			} else {
				c.verify(body, t)
			}
			for k, v := range c.resHeaders {
				if gv := res.Header.Get(k); v != gv {
					t.Fatalf("got header=%v, want header=%v", gv, v)
				}
			}
		})
	}
}
