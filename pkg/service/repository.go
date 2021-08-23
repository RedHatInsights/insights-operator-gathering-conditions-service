package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type RepositoryInterface interface {
	Rules() (*Rules, error)
}

type Repository struct {
	path string
}

func NewRepository(path string) *Repository {
	return &Repository{
		path,
	}
}

func (r Repository) readFile(filePath string) ([]byte, error) {
	f, err := os.Open(fmt.Sprintf("%s/%s", r.path, filePath))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	byteValue, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return byteValue, nil
}

func (r *Repository) Rules() (*Rules, error) {
	data, err := r.readFile("rules.json")
	if err != nil {
		return nil, err
	}

	var rules Rules
	err = json.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}

	return &rules, nil
}