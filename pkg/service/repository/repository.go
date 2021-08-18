package repository

type Interface interface{}

type Repository struct{}

func New() *Repository {
	return &Repository{}
}
