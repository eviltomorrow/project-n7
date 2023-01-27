package data

import (
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-repository/handler/db"
	"github.com/eviltomorrow/project-n7/lib/model"
	"github.com/eviltomorrow/project-n7/lib/mysql"
)

func TransmissionMetadata(date time.Time, pipe chan *model.Metadata) (int64, int64, int64, error) {
	var (
		offset, limit                   int64 = 0, 80
		size                                  = 50
		d                                     = date.Format("2006-01-02")
		timeout                               = 30 * time.Second
		affectedS, affectedD, affectedW int64
	)

	var (
		stocks = make([]*model.Stock, 0, size)
		days   = make([]*model.Quote, 0, size)
	)
	for md := range pipe {
		if md == nil {
			continue
		}

		stocks = append(stocks, &model.Stock{
			Code:            md.Code,
			Name:            md.Name,
			Suspend:         md.Suspend,
			CreateTimestamp: time.Now(),
		})

		day, err := AssembleQuoteDay(md, date)
		if err != nil {
			return affectedS, affectedD, affectedW, err
		}
		days = append(days, day)

		if len(stocks) == size {
			affected, err := StoreStock(stocks)
			if err != nil {
				return affectedS, affectedD, affectedW, err
			}
			affectedS += affected

			affected, err = StoreQuote(days, db.Day, d)
			if err != nil {
				return affectedS, affectedD, affectedW, err
			}
			affectedD += affected

			stocks, days = stocks[:0], days[:0]
		}
	}
	if len(stocks) != 0 {
		affected, err := StoreStock(stocks)
		if err != nil {
			return affectedS, affectedD, affectedW, err
		}
		affectedS += affected

		affected, err = StoreQuote(days, db.Day, d)
		if err != nil {
			return affectedS, affectedD, affectedW, err
		}
		affectedD += affected
	}

	if date.Weekday() == time.Friday {
		offset = 0
		for {
			stocks, err := db.StockWithSelectRange(mysql.DB, offset, limit, timeout)
			if err != nil {
				return affectedS, affectedD, affectedW, err
			}

			var weeks = make([]*model.Quote, 0, len(stocks))
			for _, stock := range stocks {
				week, err := AssembleQuoteWeek(stock.Code, date)
				if err != nil && err != ErrNoData {
					return affectedS, affectedD, affectedW, err
				}
				if err == ErrNoData {
					continue
				}
				if week != nil {
					weeks = append(weeks, week)
				}
			}

			affected, err := StoreQuote(weeks, db.Week, d)
			if err != nil {
				return affectedS, affectedD, affectedW, err
			}
			affectedW += affected

			if len(stocks) < int(limit) {
				break
			}
			offset += limit
		}
	}
	return affectedS, affectedD, affectedW, nil
}
