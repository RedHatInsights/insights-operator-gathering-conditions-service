package service

import "github.com/gorilla/mux"

type Handler struct {
	svc Interface
}

func NewHandler(svc Interface) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (s *Handler) Register(r *mux.Router) {
	r.Handle("/gathering_rules", gatheringRulesEndpoint(s.svc))
}
