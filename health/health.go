// Package health defines standard interfaces and utilities for health checks.
package health

import (
	"context"
)

type (
	// CheckFunc is quick way to define a health checker.
	CheckFunc func(context.Context) error

	// Checker provide functionality for checking health of a service.
	Checker interface {
		// CheckHealth establish health check to the target service.
		// Return error if target service cannot be reached
		// or not working properly.
		CheckHealth(ctx context.Context) error
	}
)

var (
	_ Checker = (CheckFunc)(nil)
)

// HealthCheck implements HealthChecker interface.
func (c CheckFunc) CheckHealth(ctx context.Context) error {
	return c(ctx)
}
