package service

type Interface interface {
	Rules() (*Rules, error)
}

type Service struct {
	repo RepositoryInterface
}

func New(repo RepositoryInterface) *Service {
	return &Service{
		repo,
	}
}

func (s *Service) Rules() (*Rules, error) {
	rules, err := s.repo.Rules()
	if err != nil {
		return nil, err
	}

	return rules, nil
}
