package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

func ChannelWithSelectOne(exec mysql.Exec, name string, timeout time.Duration) (*Channel, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = `select id, chat_id, name, create_timestamp, modify_timestamp from channel where name = ?`
	row := exec.QueryRowContext(ctx, _sql, name)
	if err := row.Err(); err != nil {
		return nil, err
	}

	var channel = &Channel{}
	if err := row.Scan(&channel.Id, &channel.ChatId, &channel.Name, &channel.CreateTimestamp, &channel.ModifyTimestamp); err != nil {
		return nil, err
	}
	return channel, nil
}

func ChannelWithInsertOne(exec mysql.Exec, channel *Channel, timeout time.Duration) (int64, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var (
		_sql = fmt.Sprintf(`insert into channel (%s) values (?, ?, now(), null)`, strings.Join(channelFields, ","))
		args = []interface{}{channel.ChatId, channel.Name}
	)

	result, err := exec.ExecContext(ctx, _sql, args)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func ChannelWithUpdateOne(exec mysql.Exec, name string, channel *Channel, timeout time.Duration) (int64, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var (
		_sql = `update channel set chat_id = ?, name = ?, modify_timestamp = now() where name = ?`
		args = []interface{}{channel.ChatId, channel.Name, channel.Name}
	)

	result, err := exec.ExecContext(ctx, _sql, args)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func ChannelWithInsertOrUpdateOne(exec mysql.Exec, name string, channel *Channel, timeout time.Duration) (int64, error) {
	if channel == nil {
		return 0, nil
	}

	if name == "" {
		return ChannelWithInsertOne(exec, channel, timeout)
	}

	data, err := ChannelWithSelectOne(exec, name, timeout)
	if err == sql.ErrNoRows {
		return ChannelWithInsertOne(exec, channel, timeout)
	}
	if err != nil {
		return 0, err
	}

	if data.ChatId != channel.ChatId || data.Name != channel.Name {
		return ChannelWithUpdateOne(exec, name, channel, timeout)
	}
	return 0, nil
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
