package storage

import (
	"time"

	"bb-project/internal/service"
)

type ObjectDTO struct {
	tableName struct{} `pg:"object"`
	Id       int       `pg:"o_id,use_zero"`
	LastSeen time.Time `pg:"last_seen"`
}

func ObjectToDTO(object service.Object) ObjectDTO {
	return ObjectDTO{Id: object.Id, LastSeen: object.LastSeen}
}
