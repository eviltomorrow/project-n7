package db

import (
	"encoding/json"
	"os"
	"time"
)

var SessionPath = "../db/session.db"

type Session struct {
	Username        string    `json:"username"`
	ChatID          int64     `json:"chat_id"`
	Status          string    `json:"status"`
	CreateTimestamp time.Time `json:"create_timestamp"`
	ModifyTimestamp time.Time `json:"modify_timestamp"`
}

func Read() (map[string]*Session, error) {
	buf, err := os.ReadFile(SessionPath)
	if os.IsNotExist(err) {
		return map[string]*Session{}, nil
	}
	if err != nil {
		return nil, err
	}
	var sessions map[string]*Session
	if err := json.Unmarshal(buf, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

func Write(sessions map[string]*Session) error {
	buf, err := json.Marshal(sessions)
	if err != nil {
		return err
	}
	return os.WriteFile(SessionPath, buf, 0644)
}
