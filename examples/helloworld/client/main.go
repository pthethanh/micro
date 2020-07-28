package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	pb "github.com/pthethanh/micro/examples/helloworld/helloworld"
	"google.golang.org/grpc"
)

func main() {
	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = os.Getenv("ADDRESS")
	}
	if addr == "" {
		addr = ":8000"
	}
	fmt.Println(addr)
	// GRPC
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	grpcClient := pb.NewGreeterClient(conn)
	// Set correlation id for tracing in the logs.
	//ctx := metadata.AppendToOutgoingContext(context.Background(), "X-Correlation-Id", "123-456-789-000")
	rep, err := grpcClient.SayHello(context.Background(), &pb.HelloRequest{
		Name: "Jack",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("GRPC Reply:", rep.Message)

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
		log.Fatal("status not ok, status_code=", res.StatusCode)
	}
	v := &pb.HelloReply{}
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		log.Fatal(err)
	}
	log.Println("HTTP Reply:", v.Message)

	// internal apis
	log.Println("READINESS:", getString(fmt.Sprintf("http://%s/internal/readiness", addr)))
	log.Println("LIVENESS:", getString(fmt.Sprintf("http://%s/internal/liveness", addr)))
	log.Printf("METRICS: \n%s\n", getString(fmt.Sprintf("http://%s/internal/metrics", addr)))
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
