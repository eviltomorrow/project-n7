package telegrambot

import (
	"sync"
	"time"
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
