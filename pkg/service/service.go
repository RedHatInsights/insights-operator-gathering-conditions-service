package service

type Interface interface {
}

type Service struct{}

func New() *Service {
	return &Service{}
}
