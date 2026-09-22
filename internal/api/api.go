package api

import (
	"context"
	"errors"

	"pii/internal/compute"
)

// ErrValidation is returned when the request body is invalid.
var ErrValidation = errors.New("validation error")

// ErrRateLimited is returned when the token bucket is exhausted.
var ErrRateLimited = errors.New("rate limited")

// Door is the thin entry point that validates, rate-limits and calls compute.
type Door struct {
	bucket *TokenBucket
	proc   *compute.Processor
}

// NewDoor creates a door.
func NewDoor(bucket *TokenBucket, proc *compute.Processor) *Door {
	return &Door{bucket: bucket, proc: proc}
}

// Process validates, rate-limits and processes a payload.
func (d *Door) Process(ctx context.Context, payload, payloadID string) (string, error) {
	if payload == "" || payloadID == "" {
		return "", ErrValidation
	}
	if !d.bucket.Acquire() {
		return "", ErrRateLimited
	}
	return d.proc.Process(ctx, payload, payloadID)
}
