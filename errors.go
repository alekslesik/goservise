package service

type appError string

const (
	ErrWrongState appError = "wrong application state"
	ErrMainOmitted appError = "main function is omitted"
	ErrTermTimeout appError = "termination timeout"
)

func (e appError) Error() string {
	return string(e)
}