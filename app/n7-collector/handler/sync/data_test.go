package sync

import (
	"testing"
	"time"

	"github.com/eviltomorrow/project-n7/lib/mongodb"
	"github.com/stretchr/testify/assert"
)

func TestSyncDataQuick(t *testing.T) {
	_assert := assert.New(t)

	mongodb.DSN = "mongodb://127.0.0.1:27017"
	err := mongodb.Build()
	_assert.Nil(err)

	total, ignore, err := DataQuick("sina")
	_assert.Nil(err)
	t.Logf("total: %v, ignore: %v", total, ignore)
	time.Sleep(5 * time.Second)
}

func TestSyncDataSlow(t *testing.T) {
	_assert := assert.New(t)

	RandomWait = [2]int{1, 3}
	mongodb.DSN = "mongodb://127.0.0.1:27017"
	err := mongodb.Build()
	_assert.Nil(err)

	total, ignore, err := DataSlow("sina")
	_assert.Nil(err)
	t.Logf("total: %v, ignore: %v", total, ignore)
	time.Sleep(5 * time.Second)
}
