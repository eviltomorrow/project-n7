package db

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/eviltomorrow/project-n7/lib/sqlite3"
)

func init() {
	sqlite3.DSN = "./telegram-bot.db"
	os.Remove(sqlite3.DSN)
	if err := sqlite3.Build(); err != nil {
		log.Fatal(err)
	}

	if err := SessionWithInitTable(sqlite3.DB, 10*time.Second); err != nil {
		log.Fatal(err)
	}
}
func TestSessionWithInsertOne(t *testing.T) {
	if _, err := SessionWithInsertOne(sqlite3.DB, &Session{Username: "shepard", ChatId: 123, Status: "subscribe"}, 10*time.Second); err != nil {
		log.Fatal(err)
	}
}

func TestSessionWithSelectRange(t *testing.T) {
	sessions, err := SessionWithSelectRange(sqlite3.DB, 0, 30, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range sessions {
		t.Logf("%s\r\n", s.String())
	}
}

func TestSessionWithSelectOne(t *testing.T) {
	session, err := SessionWithSelectOne(sqlite3.DB, "shepard", 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s\r\n", session.String())
}

func TestSessionWithUpdateOne(t *testing.T) {
	affected, err := SessionWithUpdateOne(sqlite3.DB, "shepard", &Session{Username: "shepard", ChatId: 456, Status: "unsubscribe"}, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("affected: %v\r\n", affected)

	session, err := SessionWithSelectOne(sqlite3.DB, "shepard", 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s\r\n", session.String())
}
