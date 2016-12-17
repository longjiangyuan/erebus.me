package model

type Error struct {
	Code   int
	Reason string
}

func (err *Error) Error() string {
	return err.Reason
}

func Fatal(code int, reason string) {
	panic(&Error{code, reason})
}
