package kuja

type Errors interface {
	Status() int
	Error() string
}

type DefaultErrors struct {
	status  int
	message string
}

func (e *DefaultErrors) Status() int {
	return e.status
}

func (e *DefaultErrors) Error() string {
	return e.message
}

func Error(status int, message string) error {
	return &DefaultErrors{status, message}
}
