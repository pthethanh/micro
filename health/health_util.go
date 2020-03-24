package health

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// HealthCheck return a CheckFunc for checking if the target server is OK or not.
// The target server can be considered OK, if its return status code is "200 OK".
// If the given addr is in simple form of "server:port", it will be transformed to
// http://server:port/internal/liveness
func HealthCheck(name string, addr string) CheckFunc {
	target := addr
	if !strings.HasPrefix(addr, "http") {
		target = "http://" + addr + "/internal/liveness"
	}
	return CheckFunc(func(context.Context) error {
		rs, err := http.Get(target)
		if err != nil {
			return fmt.Errorf("%s: NOK, err: %v", name, err)
		}
		defer rs.Body.Close()
		if rs.StatusCode == http.StatusOK {
			return nil
		}
		// log the detail for debugging
		b, err := ioutil.ReadAll(rs.Body)
		if err != nil {
			return fmt.Errorf("%s: NOK, err: %v", name, err)
		}
		response := string(b)
		if strings.ToLower(response) != "ok" {
			return fmt.Errorf("%s: NOK, err: %v", name, err)
		}
		// should not happen
		return fmt.Errorf("%s: NOK, status_code: %d", name, rs.StatusCode)
	})
}
