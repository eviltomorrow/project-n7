package telegrambot

import (
	"sync"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-telegram-bot/handler/db"
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

	subscribe:   map[string]*db.Session{},
	unsubscribe: make(chan string, 32),
}

func GetLib() *accumulator {
	return lib
}

type accumulator struct {
	sync.RWMutex

	syncInterval time.Duration
	stop         chan struct{}

	subscribe   map[string]*db.Session
	unsubscribe chan string
}

func (a *accumulator) Get(username string) (*db.Session, bool) {
	a.RLock()
	defer a.RUnlock()

	s, ok := a.subscribe[username]
	return s, ok
}

func (a *accumulator) Set(s *db.Session) {
	a.Lock()
	defer a.Unlock()

	if _, ok := a.subscribe[s.Username]; ok {
		if s.Status == Unsubscribe {
			a.unsubscribe <- s.Username
		}
	}
	a.subscribe[s.Username] = s
}

func (a *accumulator) load() error {
	a.Lock()
	defer a.Unlock()

	sessions, err := db.Read()
	if err != nil {
		zlog.Error("Panic: read db failure", zap.Error(err))
	} else {
		a.subscribe = sessions
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
			a.Lock()
			delete(a.subscribe, s)
			a.Unlock()
		case <-a.stop:
			return
		}
	}
}

func (a *accumulator) Flush() error {
	a.Lock()
	defer a.Unlock()

	return db.Write(a.subscribe)
}
