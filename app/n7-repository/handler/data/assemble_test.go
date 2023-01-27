package data

import (
	"testing"
	"time"

	"github.com/eviltomorrow/project-n7/lib/mysql"
	"github.com/stretchr/testify/assert"
)

func TestAssembleQuoteWeek(t *testing.T) {
	_assert := assert.New(t)
	mysql.DSN = "root:root@tcp(127.0.0.1:3306)/n7_repository?charset=utf8mb4&parseTime=true&loc=Local"
	if err := mysql.Build(); err != nil {
		t.Fatal(err)
	}

	d, err := time.Parse("2006-01-02", "2023-01-20")
	if err != nil {
		t.Fatal(err)
	}
	quote, err := AssembleQuoteWeek("sz300999", d)
	_assert.Nil(err)
	t.Logf("data: %s\r\n", quote)
}
