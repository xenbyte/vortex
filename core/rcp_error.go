package core

import "fmt"

type RPCError struct {
	Code int
	Text string
}

const (
	Timeout                = 0
	NotSupported           = 10
	TemporarilyUnavailable = 11
	MalformedRequest       = 12
	Crash                  = 13
	Abort                  = 14
	KeyDoesNotExist        = 20
	KeyAlreadyExists       = 21
	PreconditionFailed     = 22
	TxnConflict            = 30
)

func ErrorCodeText(code int) string {
	switch code {
	case Timeout:
		return "Timeout"
	case NotSupported:
		return "NotSupported"
	case TemporarilyUnavailable:
		return "TemporarilyUnavailable"
	case MalformedRequest:
		return "MalformedRequest"
	case Crash:
		return "Crash"
	case Abort:
		return "Abort"
	case KeyDoesNotExist:
		return "KeyDoesNotExist"
	case KeyAlreadyExists:
		return "KeyAlreadyExists"
	case PreconditionFailed:
		return "PreconditionFailed"
	case TxnConflict:
		return "TxnConflict"
	default:
		return fmt.Sprintf("ErrorCode<%d>", code)
	}
}

func NewRPCError(code int, text string) *RPCError {
	return &RPCError{
		Code: code,
		Text: text,
	}
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPCError(%s, %q)", ErrorCodeText(e.Code), e.Text)
}
