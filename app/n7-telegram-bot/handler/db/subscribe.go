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

func SubscribeWithSelectRange(exec mysql.Exec, offset, limit int64, timeout time.Duration) ([]*Subscribe, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = `select id, username, chat_id, status, create_timestamp, modify_timestamp from subscribe limit ?, ?`
	rows, err := exec.QueryContext(ctx, _sql, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscribes = make([]*Subscribe, 0, limit)
	for rows.Next() {
		var subscribe = &Subscribe{}
		if err := rows.Scan(&subscribe.Id, &subscribe.Username, &subscribe.ChatId, &subscribe.Status, &subscribe.CreateTimestamp, &subscribe.ModifyTimestamp); err != nil {
			return nil, err
		}
		subscribes = append(subscribes, subscribe)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return subscribes, nil
}

func SubscribeWithSelectOne(exec mysql.Exec, username string, timeout time.Duration) (*Subscribe, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = `select id, username, chat_id, status, create_timestamp, modify_timestamp from subscribe where username = ?`
	row := exec.QueryRowContext(ctx, _sql, username)
	if err := row.Err(); err != nil {
		return nil, err
	}

	var subscribe = &Subscribe{}
	if err := row.Scan(&subscribe.Id, &subscribe.Username, &subscribe.ChatId, &subscribe.Status, &subscribe.CreateTimestamp, &subscribe.ModifyTimestamp); err != nil {
		return nil, err
	}
	return subscribe, nil
}

func SubscribeWithInsertOne(exec mysql.Exec, subscribe *Subscribe, timeout time.Duration) (int64, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var (
		_sql = fmt.Sprintf(`insert into subscribe (%s) values (?, ?, ?, now(), null)`, strings.Join(SubscribeFields, ","))
		args = []interface{}{subscribe.Username, subscribe.ChatId, subscribe.Status}
	)

	result, err := exec.ExecContext(ctx, _sql, args)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func SubscribeWithUpdateOne(exec mysql.Exec, username string, subscribe *Subscribe, timeout time.Duration) (int64, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var (
		_sql = `update subscribe set username = ?, chat_id = ?, status = ?, modify_timestamp = now() where username = ?`
		args = []interface{}{subscribe.Username, subscribe.ChatId, subscribe.Status, username}
	)

	result, err := exec.ExecContext(ctx, _sql, args)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func SubscribeWithInsertOrUpdateOne(exec mysql.Exec, username string, subscribe *Subscribe, timeout time.Duration) (int64, error) {
	if subscribe == nil {
		return 0, nil
	}

	if username == "" {
		return SubscribeWithInsertOne(exec, subscribe, timeout)
	}

	data, err := SubscribeWithSelectOne(exec, username, timeout)
	if err == sql.ErrNoRows {
		return SubscribeWithInsertOne(exec, subscribe, timeout)
	}
	if err != nil {
		return 0, err
	}

	if data.ChatId != subscribe.ChatId || data.Status != subscribe.Status {
		return SubscribeWithUpdateOne(exec, username, subscribe, timeout)
	}
	return 0, nil
}

const (
	FieldSubscribeUsername    = "username"
	FieldSubscribeChatId      = "chat_id"
	FieldSubscribeStatus      = "status"
	FieldStockCreateTimestamp = "create_timestamp"
	FieldStockModifyTimestamp = "modify_timestamp"
)

var SubscribeFields = []string{
	FieldSubscribeUsername,
	FieldSubscribeChatId,
	FieldSubscribeStatus,
	FieldStockCreateTimestamp,
	FieldStockModifyTimestamp,
}

// Subscribe
type Subscribe struct {
	Id              int64        `json:"id"`
	Username        string       `json:"name"`
	ChatId          int64        `json:"chat_id"`
	Status          int          `json:"status"`
	CreateTimestamp time.Time    `json:"create_timestamp"`
	ModifyTimestamp sql.NullTime `json:"modify_timestamp"`
}

func (c *Subscribe) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(c)
	return string(buf)
}
