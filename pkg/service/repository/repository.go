package repository

type RepositoryInterface interface{}

type Repository struct{}

func New() *Repository {
	return &Repository{}
}
