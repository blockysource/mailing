package emailtemplate

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InvalidTemplateError is an error that occurs when a template is invalid.
type InvalidTemplateError struct {
	Err   error
	Field string
}

// GRPCStatus returns the gRPC status.
func (e *InvalidTemplateError) GRPCStatus() *status.Status {
	st := status.New(codes.InvalidArgument, "invalid template")
	br := errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{
			{
				Field:       e.Field,
				Description: e.Err.Error(),
			},
		},
	}
	st, _ = st.WithDetails(&br)
	return st
}

func newInvalidTemplateError(field string, err error) *InvalidTemplateError {
	return &InvalidTemplateError{
		Err:   err,
		Field: field,
	}
}

// Error returns the error message.
func (e *InvalidTemplateError) Error() string {
	return fmt.Sprintf("invalid template %s: %s", e.Field, e.Err.Error())
}

func (e *InvalidTemplateError) Unwrap() error {
	return e.Err
}

// MissingParameterError is an error that occurs when a parameter is missing.
type MissingParameterError struct {
	Parameter string
}

func newMissingParameterError(parameter string) *MissingParameterError {
	return &MissingParameterError{
		Parameter: parameter,
	}
}

func (e *MissingParameterError) Error() string {
	return fmt.Sprintf("missing parameter %s", e.Parameter)
}

// InvalidAddressError is an error that occurs when an address is invalid.
type InvalidAddressError struct {
	Err     error
	Address string
	Field   string
}

// GRPCStatus returns the gRPC status.
func (e *InvalidAddressError) GRPCStatus() *status.Status {
	st := status.New(codes.InvalidArgument, "invalid address")
	br := errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{
			{
				Field:       e.Field,
				Description: fmt.Sprintf("invalid address %s: %s", e.Address, e.Err.Error()),
			},
		},
	}
	st, _ = st.WithDetails(&br)
	return st
}

// Error returns the error message.
func (e *InvalidAddressError) Error() string {
	return fmt.Sprintf("invalid %s address %s: %s", e.Field, e.Address, e.Err.Error())
}

func newInvalidAddressError(field, address string, err error) *InvalidAddressError {
	return &InvalidAddressError{
		Err:     err,
		Field:   field,
		Address: address,
	}
}
