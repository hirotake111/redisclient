package apperror

type AppError struct {
	msg string
}

func (e *AppError) Error() string {
	return e.msg
}

var (
	CantMoveCursorDownError = &AppError{msg: "can't move cursor down"}
	CantMoveCursorUpError   = &AppError{msg: "can't move cursor up"}
)
