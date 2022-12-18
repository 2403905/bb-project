package service

import (
	"time"
)

type Object struct {
	Id       int  `json:"id"`
	Online   bool `json:"online"`
	LastSeen time.Time
}

func idToObject(id int) Object {
	return Object{Id: id}
}
