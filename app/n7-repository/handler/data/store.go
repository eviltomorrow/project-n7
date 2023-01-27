package data

import (
	"fmt"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-repository/handler/db"
	"github.com/eviltomorrow/project-n7/lib/mysql"
)

var (
	timeout = 10 * time.Second
)

func StoreStock(data []*db.Stock) (int64, error) {
	if len(data) == 0 {
		return 0, nil
	}

	tx, err := mysql.DB.Begin()
	if err != nil {
		return 0, err
	}
	affected, err := db.StockWithInsertOrUpdateMany(tx, data, timeout)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, nil
	}
	return affected, nil
}

func StoreQuote(data []*db.Quote, mode string, date string) (int64, error) {
	if len(data) == 0 {
		return 0, nil
	}
	if mode != db.Day && mode != db.Week {
		return 0, fmt.Errorf("invalid mode: %v", mode)
	}

	var codes = make([]string, 0, len(data))
	for _, d := range data {
		codes = append(codes, d.Code)
	}

	tx, err := mysql.DB.Begin()
	if err != nil {
		return 0, err
	}
	if _, err := db.QuoteWithDeleteManyByCodesAndDate(tx, mode, codes, date, timeout); err != nil {
		tx.Rollback()
		return 0, err
	}

	affected, err := db.QuoteWithInsertMany(tx, mode, data, timeout)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}
	return affected, nil
}
