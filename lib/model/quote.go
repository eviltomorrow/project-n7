package model

import (
	"database/sql"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type Quote struct {
	Id              string       `json:"id"`
	Code            string       `json:"code"`
	Open            float64      `json:"open"`
	Close           float64      `json:"close"`
	High            float64      `json:"high"`
	Low             float64      `json:"low"`
	YesterdayClosed float64      `json:"yesterday_closed"`
	Volume          uint64       `json:"volume"`
	Account         float64      `json:"account"`
	Date            time.Time    `json:"date"`
	NumOfYear       int          `json:"num_of_year"`
	Xd              float64      `json:"xd"`
	CreateTimestamp time.Time    `json:"create_timestamp"`
	ModifyTimestamp sql.NullTime `json:"modify_timestamp"`
}

func (q *Quote) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(q)
	return string(buf)
}
