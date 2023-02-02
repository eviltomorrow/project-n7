package telegrambot

import (
	"sync"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-telegram-bot/handler/db"
	"github.com/eviltomorrow/project-n7/lib/mysql"
	"github.com/eviltomorrow/project-n7/lib/zlog"
	"go.uber.org/zap"
)

type chat struct {
	id       int64
	status   int
	username string
}

type accumulator struct {
	sync.RWMutex

	SyncInterval time.Duration
	Stop         chan struct{}

	Subscribe   map[string]*chat
	Unsubscribe chan *chat
}

func (a *accumulator) get(username string) (*chat, bool) {
	a.RLock()
	defer a.RUnlock()

	s, ok := a.Subscribe[username]
	return s, ok
}

func (a *accumulator) set(s *chat) {
	a.Lock()
	defer a.Unlock()

	if _, ok := a.Subscribe[s.username]; ok {
		if s.status == 1 {
			a.Unsubscribe <- s
		}
	}
	a.Subscribe[s.username] = s
}

func (a *accumulator) load() {
	a.Lock()
	defer a.Unlock()

	var (
		offset, limit int64 = 0, 50
	)
	for {
		subscribes, err := db.SubscribeWithSelectRange(mysql.DB, offset, limit, 30*time.Second)
		if err != nil {
			zlog.Error("SubscribeWithSelectRange failure", zap.Error(err))
			break
		}
		for _, s := range subscribes {
			a.Subscribe[s.Username] = &chat{username: s.Username, id: s.ChatId, status: s.Status}
		}
		if int64(len(subscribes)) < limit {
			break
		}
		offset += limit
	}
	go a.sync()
}

func (a *accumulator) sync() {
	var ticker = time.NewTicker(a.SyncInterval)
	for {
		select {
		case <-ticker.C:
			var subscribes = make([]*chat, 0, 32)
			a.Lock()
			for _, v := range a.Subscribe {
				subscribes = append(subscribes, v)
			}
			a.Unlock()

			for _, s := range subscribes {
				if _, err := db.SubscribeWithInsertOrUpdateOne(mysql.DB, s.username, &db.Subscribe{Username: s.username, ChatId: s.id, Status: s.status}, 10*time.Second); err != nil {
					zlog.Error("SubscribeWithInsertOrUpdateOne failure", zap.Error(err))
				}
			}
		case s := <-a.Unsubscribe:
			if _, err := db.SubscribeWithUpdateOne(mysql.DB, s.username, &db.Subscribe{Username: s.username, ChatId: s.id, Status: 1}, 10*time.Second); err != nil {
				zlog.Error("SubscribeWithUpdateOne failure", zap.Error(err))
			}
		case <-a.Stop:
			return
		}
	}
}
