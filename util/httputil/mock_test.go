package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pthethanh/micro/util/httputil"
)

func TestMock(t *testing.T) {
	specs := httputil.MustReadMockHandlersFromFile("testdata/mock.json")
	wantN := 6
	if len(specs) != wantN {
		t.Fatalf("got number of spec = %d, want number of spec = %d", len(specs), wantN)
	}

	srv := httptest.NewServer(httputil.Mock(specs...))
	defer srv.Close()

	cases := []struct {
		name    string
		headers map[string]string
		method  string
		path    string

		code       int
		body       map[string]interface{}
		resHeaders map[string]string
	}{
		{
			name:   "get users ok",
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
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// TODO fix me.
			// req, err := http.NewRequest(c.method, srv.URL+c.path, nil)
			// if err != nil {
			// 	t.Fatal(err)
			// }
			// for k, v := range c.headers {
			// 	req.Header.Set(k, v)
			// }
			// res, err := http.DefaultClient.Do(req)
			// if err != nil {
			// 	t.Error(err)
			// }
			// defer res.Body.Close()
			// if res.StatusCode != c.code {
			// 	t.Fatalf("got status=%d, want status=%d", res.StatusCode, c.code)
			// }
			// var body map[string]interface{}
			// if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			// 	t.Error(err)
			// }
			// for k, v := range c.body {
			// 	if gv, ok := body[k]; !ok || !reflect.DeepEqual(v, gv) {
			// 		t.Fatalf("got body[%s]=%v, want body[%s]=%v", k, gv, k, v)
			// 	}
			// }
			// for k, v := range c.resHeaders {
			// 	if gv := res.Header.Get(k); v != gv {
			// 		t.Fatalf("got header=%v, want header=%v", gv, v)
			// 	}
			// }
		})
	}
}
