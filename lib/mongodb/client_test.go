package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	_assert := assert.New(t)
	DSN = "mongodb://admin:admin123@127.0.0.1:27017/n7"

	err := Build()
	_assert.Nil(err)
}
