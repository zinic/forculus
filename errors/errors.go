package errors

type New string

func (s New) Error() string {
	return string(s)
}
