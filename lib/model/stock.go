package model

import (
	"database/sql"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// Stock
type Stock struct {
	Code            string       `json:"code"`
	Name            string       `json:"name"`
	Suspend         string       `json:"suspend"`
	CreateTimestamp time.Time    `json:"create_timestamp"`
	ModifyTimestamp sql.NullTime `json:"modify_timestamp"`
}

func (s *Stock) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(s)
	return string(buf)
}
