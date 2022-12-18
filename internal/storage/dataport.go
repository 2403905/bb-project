package storage

import (
	"context"
	"errors"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/rs/zerolog/log"

	"bb-project/db"
	"bb-project/internal/service"
)

type DataPort struct {
	db *db.PgDatabase
}

func NewDataPort(db *db.PgDatabase) *DataPort {
	return &DataPort{db: db}
}

func (s *DataPort) SaveObjects(ctx context.Context, objectList []service.Object) error {
	dtoList := make([]ObjectDTO, len(objectList))
	for k := range objectList {
		dtoList[k] = ObjectToDTO(objectList[k])
	}
	return s.saveObjects(ctx, dtoList)
}

func (s *DataPort) saveObjects(ctx context.Context, dtoList []ObjectDTO) error {
	db, err := s.db.GetDbE()
	if err != nil {
		return err
	}
	res, err := db.ModelContext(ctx, &dtoList).OnConflict("(o_id) DO UPDATE").Insert()
	if err != nil {
		return err
	}
	log.Debug().Msgf("inserted %d objects", res.RowsAffected())
	return nil
}

func (s *DataPort) RemoveObjects(ctx context.Context, retention time.Time) error {
	db, err := s.db.GetDbE()
	if err != nil {
		return err
	}
	deleted := []int{}
	res, err := db.ModelContext(ctx, &ObjectDTO{}).
		Where("last_seen < ?", retention).
		Returning("o_id").
		Delete(&deleted)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		return err
	}
	if err == nil {
		log.Debug().Msgf("deleted %d objects: %v", res.RowsAffected(), deleted)
	}
	return nil
}
