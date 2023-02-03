package telegrambot

import (
	"sync"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-telegram-bot/handler/db"
	"github.com/eviltomorrow/project-n7/lib/mysql"
	"github.com/eviltomorrow/project-n7/lib/zlog"
	"go.uber.org/zap"
)

const (
	Subscribe   = "subscribe"
	Unsubscribe = "unsubscribe"
)

var lib = &accumulator{
	syncInterval: 2 * time.Hour,
	stop:         make(chan struct{}),

	subscribe:   map[string]*Session{},
	unsubscribe: make(chan *Session, 32),
}

func GetLib() *accumulator {
	return lib
}

type Session struct {
	Status   string
	ChatId   int64
	Username string
}
type accumulator struct {
	sync.RWMutex

	syncInterval time.Duration
	stop         chan struct{}

	subscribe   map[string]*Session
	unsubscribe chan *Session
}

func (a *accumulator) Get(username string) (*Session, bool) {
	a.RLock()
	defer a.RUnlock()

	s, ok := a.subscribe[username]
	return s, ok
}

func (a *accumulator) Set(s *Session) {
	a.Lock()
	defer a.Unlock()

	if _, ok := a.subscribe[s.Username]; ok {
		if s.Status == Unsubscribe {
			a.unsubscribe <- s
		}
	}
	a.subscribe[s.Username] = s
}

func (a *accumulator) load() error {
	a.Lock()
	defer a.Unlock()

	var (
		offset, limit int64 = 0, 50
	)
	for {
		sessions, err := db.SessionWithSelectRange(mysql.DB, offset, limit, 30*time.Second)
		if err != nil {
			return err
		}
		for _, s := range sessions {
			a.subscribe[s.Username] = &Session{Username: s.Username, ChatId: s.ChatId, Status: s.Status}
		}
		if int64(len(sessions)) < limit {
			break
		}
		offset += limit
	}
	go a.sync()

	return nil
}

func (a *accumulator) sync() {
	var ticker = time.NewTicker(a.syncInterval)
	for {
		select {
		case <-ticker.C:
			if err := a.Flush(); err != nil {
				zlog.Error("Flush session data failure", zap.Error(err))
			}
		case s := <-a.unsubscribe:
			if _, err := db.SessionWithUpdateOne(mysql.DB, s.Username, &db.Session{Username: s.Username, ChatId: s.ChatId, Status: Unsubscribe}, 10*time.Second); err != nil {
				zlog.Error("SessionWithUpdateOne failure", zap.Error(err))
			}
			a.Lock()
			delete(a.subscribe, s.Username)
			a.Unlock()
		case <-a.stop:
			return
		}
	}
}

func (a *accumulator) Flush() error {
	var sessions = make([]*Session, 0, 32)
	a.Lock()
	for _, v := range a.subscribe {
		sessions = append(sessions, v)
	}
	a.Unlock()

	for _, s := range sessions {
		if _, err := db.SessionWithInsertOrUpdateOne(mysql.DB, s.Username, &db.Session{Username: s.Username, ChatId: s.ChatId, Status: s.Status}, 10*time.Second); err != nil {
			return err
		}
	}
	return nil
}
