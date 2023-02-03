package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/eviltomorrow/project-n7/lib/sqlite3"
	jsoniter "github.com/json-iterator/go"
)

func SessionWithInitTable(exec sqlite3.Exec, timeout time.Duration) error {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = `CREATE TABLE IF NOT EXISTS session (
id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
username VARCHAR(32) NOT NULL UNIQUE,
chat_id INTEGER NOT NULL,
status VARCHAR(16) NOT NULL,
create_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
modify_timestamp TIMESTAMP
);`
	if _, err := exec.ExecContext(ctx, _sql, nil); err != nil {
		return err
	}
	return nil
}

func SessionWithSelectRange(exec sqlite3.Exec, offset, limit int64, timeout time.Duration) ([]*Session, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = `select id, username, chat_id, status, create_timestamp, modify_timestamp from session limit ?, ?`
	rows, err := exec.QueryContext(ctx, _sql, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions = make([]*Session, 0, limit)
	for rows.Next() {
		var session = &Session{}
		if err := rows.Scan(&session.Id, &session.Username, &session.ChatId, &session.Status, &session.CreateTimestamp, &session.ModifyTimestamp); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func SessionWithSelectOne(exec sqlite3.Exec, username string, timeout time.Duration) (*Session, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = `select id, username, chat_id, status, create_timestamp, modify_timestamp from session where username = ?`
	row := exec.QueryRowContext(ctx, _sql, username)
	if err := row.Err(); err != nil {
		return nil, err
	}

	var session = &Session{}
	if err := row.Scan(&session.Id, &session.Username, &session.ChatId, &session.Status, &session.CreateTimestamp, &session.ModifyTimestamp); err != nil {
		return nil, err
	}
	return session, nil
}

func SessionWithInsertOne(exec sqlite3.Exec, session *Session, timeout time.Duration) (int64, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var (
		_sql = fmt.Sprintf(`insert into session (%s) values (?, ?, ?, DateTime('now', 'localtime'), null)`, strings.Join(SessionFields, ","))
		args = []interface{}{session.Username, session.ChatId, session.Status}
	)

	result, err := exec.ExecContext(ctx, _sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func SessionWithUpdateOne(exec sqlite3.Exec, username string, session *Session, timeout time.Duration) (int64, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var (
		_sql = `update session set chat_id = ?, status = ?, modify_timestamp = DateTime('now', 'localtime') where username = ?`
		args = []interface{}{session.ChatId, session.Status, username}
	)

	result, err := exec.ExecContext(ctx, _sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func SessionWithInsertOrUpdateOne(exec sqlite3.Exec, username string, session *Session, timeout time.Duration) (int64, error) {
	if session == nil {
		return 0, nil
	}

	if username == "" {
		return SessionWithInsertOne(exec, session, timeout)
	}

	data, err := SessionWithSelectOne(exec, username, timeout)
	if err == sql.ErrNoRows {
		return SessionWithInsertOne(exec, session, timeout)
	}
	if err != nil {
		return 0, err
	}

	if data.ChatId != session.ChatId || data.Status != session.Status {
		return SessionWithUpdateOne(exec, username, session, timeout)
	}
	return 0, nil
}

const (
	FieldSessionUsername      = "username"
	FieldSessionChatId        = "chat_id"
	FieldSessionStatus        = "status"
	FieldStockCreateTimestamp = "create_timestamp"
	FieldStockModifyTimestamp = "modify_timestamp"
)

var SessionFields = []string{
	FieldSessionUsername,
	FieldSessionChatId,
	FieldSessionStatus,
	FieldStockCreateTimestamp,
	FieldStockModifyTimestamp,
}

// Session
type Session struct {
	Id              int64        `json:"id"`
	Username        string       `json:"name"`
	ChatId          int64        `json:"chat_id"`
	Status          string       `json:"status"`
	CreateTimestamp time.Time    `json:"create_timestamp"`
	ModifyTimestamp sql.NullTime `json:"modify_timestamp"`
}

func (c *Session) String() string {
	buf, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(c)
	return string(buf)
}
