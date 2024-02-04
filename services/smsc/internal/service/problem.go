package service

type TransactionProblem interface {
	error
	GetTitle() string
	GetDetail() string
	GetStatusCode() int
	GetErrorType() string
}
