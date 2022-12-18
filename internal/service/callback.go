package service

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

type Callback struct {
	client  *http.Client
	produce func(message string)
}

func NewCallback(produce func(message string)) *Callback {
	s := &Callback{
		client:  http.DefaultClient,
		produce: produce,
	}
	return s
}

func (s *Callback) Callback(ids []int) {
	b, err := json.Marshal(ids)
	if err != nil {
		log.Err(err).Send()
	}
	go s.produce(string(b))
}
