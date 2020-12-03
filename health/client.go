package health

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// NewClient return a new HealthClient.
func NewClient(cc grpc.ClientConnInterface) grpc_health_v1.HealthClient {
	return grpc_health_v1.NewHealthClient(cc)
}
