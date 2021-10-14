package service

import (
	"encoding/json"
	"fmt"
)

type RepositoryInterface interface {
	Rules() (*Rules, error)
}

type Rule struct {
	Conditions         []interface{} `json:"conditions,omitempty"`
	GatheringFunctions interface{}   `json:"gathering_functions,omitempty"`
}

type Rules struct {
	Items []Rule `json:"rules,omitempty"`
}

type Repository struct {
	store StorageInterface
}

func NewRepository(s StorageInterface) *Repository {
	return &Repository{store: s}
}

func (r *Repository) Rules() (*Rules, error) {
	filepath := "rules.json"
	data := r.store.Find(filepath)
	if data == nil {
		return nil, fmt.Errorf("store data not found for '%s'", filepath)
	}

	var rules Rules
	err := json.Unmarshal(data, &rules)
	if err != nil {
		return nil, err
	}

	return &rules, nil
}
