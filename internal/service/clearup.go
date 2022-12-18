package service

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type ClearUpDataPort interface {
	RemoveObjects(ctx context.Context, retention time.Time) error
}

type ClearUp struct {
	data   ClearUpDataPort
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewClearUp(dataPort ClearUpDataPort) *ClearUp {
	ctx, cancel := context.WithCancel(context.Background())
	return &ClearUp{data: dataPort, ctx: ctx, cancel: cancel}
}

func (s *ClearUp) Run() {
	go s.run()
}

func (s *ClearUp) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *ClearUp) run() {
	s.wg.Add(1)
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			log.Debug().Msg("clear up loop stopped")
			return
		case tick := <-time.After(1 * time.Second):
			retention := time.Now().Add(-30 * time.Second).UTC()
			log.Debug().Msgf("clear up flush period reached %s", tick)
			err := s.data.RemoveObjects(s.ctx, retention)
			if err != nil {
				log.Err(err).Msg("object removing error")
			}
		}
	}
}
