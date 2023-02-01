package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/eviltomorrow/project-n7/lib/mysql"
	jsoniter "github.com/json-iterator/go"
)

func ChannelWithSelectRange(exec mysql.Exec, offset, limit int64, timeout time.Duration) ([]*Channel, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = `select id, chat_id, name, create_timestamp, modify_timestamp from channel limit ?, ?`
	rows, err := exec.QueryContext(ctx, _sql, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels = make([]*Channel, 0, limit)
	for rows.Next() {
		var channel = &Channel{}
		if err := rows.Scan(&channel.Id, &channel.ChatId, &channel.Name, &channel.CreateTimestamp, &channel.ModifyTimestamp); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return channels, nil
}

const (
	FieldChannelChatId        = "chat_id"
	FieldChannelName          = "name"
	FieldStockCreateTimestamp = "create_timestamp"
	FieldStockModifyTimestamp = "modify_timestamp"
)

var channelFields = []string{
	FieldChannelChatId,
	FieldChannelName,
	FieldStockCreateTimestamp,
	FieldStockModifyTimestamp,
}

// Channel
type Channel struct {
	Id              int64        `json:"id"`
	ChatId          int64        `json:"chat_id"`
	Name            string       `json:"name"`
	CreateTimestamp time.Time    `json:"create_timestamp"`
	ModifyTimestamp sql.NullTime `json:"modify_timestamp"`
}

func (c *Channel) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(c)
	return string(buf)
}
