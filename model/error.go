package model

const (
	// ErrSmthWentWrong error smth went wrong.
	ErrSmthWentWrong = Error("smth went wrong")
)

type Error string

func (e Error) Error() string {
	return string(e)
}
