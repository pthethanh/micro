// Package status provide wrapper of package status, codes of grpc-go and some utilities for working with status, errors.
package status

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pthethanh/micro/log"

	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// Status is an alias of grpc status.Status.
	Status = status.Status

	// Code is an alias of grpc codes.Code.
	Code = codes.Code
)

// New return new status.
func New(code Code, fmt string, args ...interface{}) *Status {
	if len(args) == 0 {
		if fmt == "" {
			return status.New(code, code.String())
		}
		return status.New(code, fmt)
	}
	return status.Newf(code, fmt, args...)
}

// Error return new error.
func Error(code Code, fmt string, args ...interface{}) error {
	if len(args) == 0 {
		if fmt == "" {
			return status.Error(code, code.String())
		}
		return status.Error(code, fmt)
	}
	return status.Errorf(code, fmt, args...)
}

// Is check whether the given error has same code with the given code.
func Is(err error, code Code) bool {
	return status.Code(err) == code
}

// Convert is a convenience function which removes the need to handle the
// boolean return value from FromError.
func Convert(err error) *Status {
	return status.Convert(err)
}

// FromError returns a Status representing err if it was produced from this
// package or has a method `GRPCStatus() *Status`. Otherwise, ok is false and a
// Status is returned with codes.Unknown and the original error message.
func FromError(err error) (s *Status, ok bool) {
	return status.FromError(err)
}

// OK is returned on success.
func OK(fmt string, args ...interface{}) *Status {
	return New(codes.OK, fmt, args...)
}

// IsOK check whether the given error is actually OK message.
func IsOK(err error) bool {
	return Is(err, codes.OK)
}

// Canceled indicates the operation was canceled (typically by the caller).
func Canceled(fmt string, args ...interface{}) error {
	return Error(codes.Canceled, fmt, args...)
}

// IsCanceled check whether the error is a Canceled error.
func IsCanceled(err error) bool {
	return Is(err, codes.Canceled)
}

// Unknown error. An example of where this error may be returned is
// if a Status value received from another address space belongs to
// an error-space that is not known in this address space. Also
// errors raised by APIs that do not return enough error information
// may be converted to this error.
func Unknown(fmt string, args ...interface{}) error {
	return Error(codes.Unknown, fmt, args...)
}

// IsUnknown check whether the given error is Unknown error.
func IsUnknown(err error) bool {
	return Is(err, codes.Unknown)
}

// InvalidArgument indicates client specified an invalid argument.
// Note that this differs from FailedPrecondition. It indicates arguments
// that are problematic regardless of the state of the system
// (e.g., a malformed file name).
func InvalidArgument(fmt string, args ...interface{}) error {
	return Error(codes.InvalidArgument, fmt, args...)
}

// IsInvalidArgument check whether the given error is an InvalidArgument.
func IsInvalidArgument(err error) bool {
	return Is(err, codes.InvalidArgument)
}

// DeadlineExceeded means operation expired before completion.
// For operations that change the state of the system, this error may be
// returned even if the operation has completed successfully. For
// example, a successful response from a server could have been delayed
// long enough for the deadline to expire.
func DeadlineExceeded(fmt string, args ...interface{}) error {
	return Error(codes.DeadlineExceeded, fmt, args...)
}

// IsDeadlineExceeded check whether the given error is a DeadlineExceeded error.
func IsDeadlineExceeded(err error) bool {
	return Is(err, codes.DeadlineExceeded)
}

// NotFound means some requested entity (e.g., file or directory) was
// not found.
func NotFound(fmt string, args ...interface{}) error {
	return Error(codes.NotFound, fmt, args...)
}

// IsNotFound check whether the given error is a NotFound error.
func IsNotFound(err error) bool {
	return Is(err, codes.NotFound)
}

// AlreadyExists means an attempt to create an entity failed because one
// already exists.
func AlreadyExists(fmt string, args ...interface{}) error {
	return Error(codes.AlreadyExists, fmt, args...)
}

// IsAlreadyExists check whether the given error is an AlreadyExists error.
func IsAlreadyExists(err error) bool {
	return Is(err, codes.AlreadyExists)
}

// PermissionDenied indicates the caller does not have permission to
// execute the specified operation. It must not be used for rejections
// caused by exhausting some resource (use ResourceExhausted
// instead for those errors). It must not be
// used if the caller cannot be identified (use Unauthenticated
// instead for those errors).
func PermissionDenied(fmt string, args ...interface{}) error {
	return Error(codes.PermissionDenied, fmt, args...)
}

// IsPermissionDenied check whether the given error is a PermissionDenied error.
func IsPermissionDenied(err error) bool {
	return Is(err, codes.PermissionDenied)
}

// ResourceExhausted indicates some resource has been exhausted, perhaps
// a per-user quota, or perhaps the entire file system is out of space.
//
// This error code will be generated by the gRPC framework in
// out-of-memory and server overload situations, or when a message is
// larger than the configured maximum size.
func ResourceExhausted(fmt string, args ...interface{}) error {
	return Error(codes.ResourceExhausted, fmt, args...)
}

// IsResourceExhausted check whether the given error is a ResourceExhausted error.
func IsResourceExhausted(err error) bool {
	return Is(err, codes.ResourceExhausted)
}

// FailedPrecondition indicates operation was rejected because the
// system is not in a state required for the operation's execution.
// For example, directory to be deleted may be non-empty, an rmdir
// operation is applied to a non-directory, etc.
//
// A litmus test that may help a service implementor in deciding
// between FailedPrecondition, Aborted, and Unavailable:
//  (a) Use Unavailable if the client can retry just the failing call.
//  (b) Use Aborted if the client should retry at a higher-level
//      (e.g., restarting a read-modify-write sequence).
//  (c) Use FailedPrecondition if the client should not retry until
//      the system state has been explicitly fixed. E.g., if an "rmdir"
//      fails because the directory is non-empty, FailedPrecondition
//      should be returned since the client should not retry unless
//      they have first fixed up the directory by deleting files from it.
//  (d) Use FailedPrecondition if the client performs conditional
//      REST Get/Update/Delete on a resource and the resource on the
//      server does not match the condition. E.g., conflicting
//      read-modify-write on the same resource.
func FailedPrecondition(fmt string, args ...interface{}) error {
	return Error(codes.FailedPrecondition, fmt, args...)
}

// IsFailedPrecondition check whether the given error is a FailedPrecondition error.
func IsFailedPrecondition(err error) bool {
	return Is(err, codes.FailedPrecondition)
}

// Aborted indicates the operation was aborted, typically due to a
// concurrency issue like sequencer check failures, transaction aborts,
// etc.
func Aborted(fmt string, args ...interface{}) error {
	return Error(codes.Aborted, fmt, args...)
}

// IsAborted check whether the given error is an Aborted error.
func IsAborted(err error) bool {
	return Is(err, codes.Aborted)
}

// OutOfRange means operation was attempted past the valid range.
// E.g., seeking or reading past end of file.
//
// Unlike InvalidArgument, this error indicates a problem that may
// be fixed if the system state changes. For example, a 32-bit file
// system will generate InvalidArgument if asked to read at an
// offset that is not in the range [0,2^32-1], but it will generate
// OutOfRange if asked to read from an offset past the current
// file size.
//
// There is a fair bit of overlap between FailedPrecondition and
// OutOfRange. We recommend using OutOfRange (the more specific
// error) when it applies so that callers who are iterating through
// a space can easily look for an OutOfRange error to detect when
// they are done.
func OutOfRange(fmt string, args ...interface{}) error {
	return Error(codes.OutOfRange, fmt, args...)
}

// IsOutOfRange check whether the given error is an OutOfRange error.
func IsOutOfRange(err error) bool {
	return Is(err, codes.OutOfRange)
}

// Unimplemented indicates operation is not implemented or not
// supported/enabled in this service.
//
// This error code will be generated by the gRPC framework. Most
// commonly, you will see this error code when a method implementation
// is missing on the server. It can also be generated for unknown
// compression algorithms or a disagreement as to whether an RPC should
// be streaming.
func Unimplemented(fmt string, args ...interface{}) error {
	return Error(codes.Unimplemented, fmt, args...)
}

// IsUnimplemented check whether the given error is an Unimplemented error.
func IsUnimplemented(err error) bool {
	return Is(err, codes.Unimplemented)
}

// Internal errors. Means some invariants expected by underlying
// system has been broken. If you see one of these errors,
// something is very broken.
func Internal(fmt string, args ...interface{}) error {
	return Error(codes.Internal, fmt, args...)
}

// IsInternal check whether the given error is an Internal error.
func IsInternal(err error) bool {
	return Is(err, codes.Internal)
}

// Unavailable indicates the service is currently unavailable.
// This is a most likely a transient condition and may be corrected
// by retrying with a backoff. Note that it is not always safe to retry
// non-idempotent operations.
func Unavailable(fmt string, args ...interface{}) error {
	return Error(codes.Unavailable, fmt, args...)
}

// IsUnavailable check whether the given error is an Unavailable error.
func IsUnavailable(err error) bool {
	return Is(err, codes.Unavailable)
}

// DataLoss indicates unrecoverable data loss or corruption.
func DataLoss(fmt string, args ...interface{}) error {
	return Error(codes.DataLoss, fmt, args...)
}

// IsDataLoss check whether the given error is a DataLoss error.
func IsDataLoss(err error) bool {
	return Is(err, codes.DataLoss)
}

// Unauthenticated indicates the request does not have valid
// authentication credentials for the operation.
func Unauthenticated(fmt string, args ...interface{}) error {
	return Error(codes.Unauthenticated, fmt, args...)
}

// IsUnauthenticated check whether the given error is an Unauthenticated error.
func IsUnauthenticated(err error) bool {
	return Is(err, codes.Unauthenticated)
}

// HTTPStatusFromCode converts a gRPC error code into the corresponding HTTP response status.
// See: https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func HTTPStatusFromCode(code Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		// Note, this deliberately doesn't translate to the similarly named '412 Precondition Failed' HTTP response status.
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}

// JSON return JSON encoded of the given error.
func JSON(err error) []byte {
	b, mErr := json.Marshal(Convert(err).Proto())
	if mErr != nil {
		log.Errorf("errors: marshall error, err: %v", mErr)
		return []byte(fmt.Sprintf(`{"code":%d, "message":"%s"}`, codes.Internal, codes.Internal.String()))
	}
	return b
}

// Parse try to parse the given data to a Status.
func Parse(data []byte) (*Status, error) {
	s := spb.Status{}
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return status.FromProto(&s), nil
}
