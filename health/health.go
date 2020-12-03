// Package health defines standard interfaces and utilities for health checks.
package health

import (
	"context"
)

// CheckFunc function signature for health checks.
type CheckFunc func(context.Context) error
