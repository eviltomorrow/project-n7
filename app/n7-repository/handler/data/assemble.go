package data

import (
	"errors"
	"time"

	"github.com/eviltomorrow/project-n7/app/n7-repository/handler/db"
	"github.com/eviltomorrow/project-n7/lib/mathutil"
	"github.com/eviltomorrow/project-n7/lib/model"
	"github.com/eviltomorrow/project-n7/lib/mysql"
	"github.com/eviltomorrow/project-n7/lib/snowflake"
	"github.com/eviltomorrow/project-n7/lib/timeutil"
)

var (
	ErrNoData = errors.New("no data")
)

func AssembleQuoteDay(data *model.Metadata, date time.Time) (*model.Quote, error) {
	latest, err := db.QuoteWithSelectManyLatest(mysql.DB, db.Day, data.Code, data.Date, 1, timeout)
	if err != nil {
		return nil, err
	}

	var xd float64 = 1.0
	if len(latest) == 1 && latest[0].Close != 0 && latest[0].Date.Format("2006-01-02") != data.Date && latest[0].Close != data.YesterdayClosed {
		xd = data.YesterdayClosed / latest[0].Close
	}

	quote := &model.Quote{
		Id:              snowflake.Generate().String(),
		Code:            data.Code,
		Open:            data.Open,
		Close:           data.Latest,
		High:            data.High,
		Low:             data.Low,
		YesterdayClosed: data.YesterdayClosed,
		Volume:          data.Volume,
		Account:         data.Account,
		Date:            date,
		NumOfYear:       date.YearDay(),
		Xd:              xd,
		CreateTimestamp: time.Now(),
	}
	return quote, nil
}

func AssembleQuoteWeek(code string, date time.Time) (*model.Quote, error) {
	var (
		begin = date.AddDate(0, 0, -5).Format("2006-01-02")
		end   = date.Format("2006-01-02")
	)

	days, err := db.QuoteWithSelectBetweenByCodeAndDate(mysql.DB, db.Day, code, begin, end, timeout)
	if err != nil {
		return nil, err
	}

	if len(days) == 0 {
		return nil, ErrNoData
	}

	var (
		first, last = days[0], days[len(days)-1]
		highs       = make([]float64, 0, len(days))
		lows        = make([]float64, 0, len(days))
		volumes     = make([]uint64, 0, len(days))
		accounts    = make([]float64, 0, len(days))
	)

	var xd = 1.0
	for _, d := range days {
		highs = append(highs, d.High)
		lows = append(lows, d.Low)
		volumes = append(volumes, d.Volume)
		accounts = append(accounts, d.Account)
		if d.Xd != 1.0 {
			xd = d.Xd
		}
	}

	var week = &model.Quote{
		Id:              snowflake.Generate().String(),
		Code:            first.Code,
		Open:            first.Open,
		Close:           last.Close,
		High:            mathutil.Max(highs),
		Low:             mathutil.Min(lows),
		YesterdayClosed: first.YesterdayClosed,
		Volume:          mathutil.Sum(volumes),
		Account:         mathutil.Sum(accounts),
		Date:            date,
		NumOfYear:       timeutil.YearWeek(date),
		Xd:              xd,
		CreateTimestamp: time.Now(),
	}
	return week, nil
}
