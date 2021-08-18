package service

type Interface interface {
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}
