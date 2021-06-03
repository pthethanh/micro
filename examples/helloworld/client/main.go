package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pthethanh/micro/client"
	pb "github.com/pthethanh/micro/examples/helloworld/helloworld"
	"github.com/pthethanh/micro/health"
	_ "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	addr := client.GetAddressFromEnv()
	conn := client.Must(client.Dial(addr))
	client := pb.NewGreeterClient(conn)
	// Set correlation id for tracing in the logs.
	//ctx := metadata.AppendToOutgoingContext(context.Background(), "X-Correlation-Id", "123-456-789-000")
	rep, err := client.SayHello(context.Background(), &pb.HelloRequest{
		Name: "Jack",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("RESPONSE GRPC:", rep.Message)

	// HTTP
	body := bytes.NewBuffer([]byte(`{"name":"Jack"}`))
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/api/v1/hello", addr), body)
	if err != nil {
		log.Fatal(err)
	}
	// Set correlation id for tracing in the logs.
	// req.Header.Set("X-Correlation-Id", "123-456-789-001")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Fatalf("got status_code=%d, want status_code=%d", res.StatusCode, http.StatusOK)
	}
	v := &pb.HelloReply{}
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		log.Fatal(err)
	}
	log.Println("RESPONSE HTTP:", v.Message)

	// internal apis
	log.Println("HEALTH CHECK HTTP:", getString(fmt.Sprintf("http://%s/internal/health", addr)))
	log.Printf("METRICS: \n%s\n", getString(fmt.Sprintf("http://%s/internal/metrics", addr)))

	rs, err := health.NewClient(conn).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{
		Service: "",
	})
	if err != nil {
		log.Fatal("health check failed", err)
	}
	if rs.Status != health.StatusServing {
		log.Fatalf("got health status=%d, want status=%d", rs.Status, health.StatusServing)
	}
	log.Println("HEALTH CHECK GRPC:", rs.Status)
}

func getString(url string) string {
	res, err := http.DefaultClient.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Fatal("status code not ok, status_code=", res.StatusCode)
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}
