package service

type appError string

const (
	ErrWrongState appError = "wrong application state"
)

func (e appError) Error() string {
	return string(e)
}