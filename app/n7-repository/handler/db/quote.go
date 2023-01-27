package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eviltomorrow/project-n7/lib/mathutil"
	"github.com/eviltomorrow/project-n7/lib/model"
	"github.com/eviltomorrow/project-n7/lib/mysql"
)

const (
	Day  = "day"
	Week = "week"
)

func QuoteWithInsertMany(exec mysql.Exec, model string, data []*model.Quote, timeout time.Duration) (int64, error) {
	if len(data) == 0 {
		return 0, nil
	}

	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var FieldQuotes = make([]string, 0, len(data))
	var args = make([]interface{}, 0, 11*len(data))
	for _, m := range data {
		FieldQuotes = append(FieldQuotes, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now())")
		args = append(args, m.Code)
		args = append(args, m.Open)
		args = append(args, m.Close)
		args = append(args, m.High)
		args = append(args, m.Low)
		args = append(args, m.YesterdayClosed)
		args = append(args, m.Volume)
		args = append(args, m.Account)
		args = append(args, m.Date.Format("2006-01-02"))
		args = append(args, m.NumOfYear)
		args = append(args, m.Xd)
	}

	var _sql = fmt.Sprintf("insert into quote_%s (%s) values %s", model, strings.Join(quoteFeilds, ","), strings.Join(FieldQuotes, ","))
	result, err := exec.ExecContext(ctx, _sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func QuoteWithDeleteManyByCodesAndDate(exec mysql.Exec, model string, codes []string, date string, timeout time.Duration) (int64, error) {
	if len(codes) == 0 {
		return 0, nil
	}

	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var FieldQuotes = make([]string, 0, len(codes))
	var args = make([]interface{}, 0, len(codes)+1)
	for _, code := range codes {
		FieldQuotes = append(FieldQuotes, "?")
		args = append(args, code)
	}
	args = append(args, date)

	var _sql = fmt.Sprintf("delete from quote_%s where code in (%s) and DATE_FORMAT(`date`, '%%Y-%%m-%%d') = ?", model, strings.Join(FieldQuotes, ","))
	result, err := exec.ExecContext(ctx, _sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func QuoteWithSelectBetweenByCodeAndDate(exec mysql.Exec, mode string, code string, begin, end string, timeout time.Duration) ([]*model.Quote, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = fmt.Sprintf("select id, code, open, close, high, low, yesterday_closed, volume, account, date, num_of_year, xd, create_timestamp, modify_timestamp from quote_%s where code = ? and DATE_FORMAT(`date`, '%%Y-%%m-%%d') between ? and ? order by date asc", mode)
	rows, err := exec.QueryContext(ctx, _sql, code, begin, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data = make([]*model.Quote, 0, 5)
	for rows.Next() {
		var m = model.Quote{}
		if err := rows.Scan(
			&m.Id,
			&m.Code,
			&m.Open,
			&m.Close,
			&m.High,
			&m.Low,
			&m.YesterdayClosed,
			&m.Volume,
			&m.Account,
			&m.Date,
			&m.NumOfYear,
			&m.Xd,
			&m.CreateTimestamp,
			&m.ModifyTimestamp,
		); err != nil {
			return nil, err
		}
		data = append(data, &m)
	}

	var result = make([]*model.Quote, len(data))
	var xd float64 = 1.0
	for i := len(data) - 1; i >= 0; i-- {
		var d = data[i]
		if xd != 1.0 {
			var n = &model.Quote{
				Id:              d.Id,
				Code:            d.Code,
				Open:            mathutil.Trunc2(d.Open * xd),
				Close:           mathutil.Trunc2(d.Close * xd),
				High:            mathutil.Trunc2(d.High * xd),
				Low:             mathutil.Trunc2(d.Low * xd),
				YesterdayClosed: mathutil.Trunc2(d.YesterdayClosed * xd),
				Volume:          d.Volume,
				Account:         d.Account,
				Date:            d.Date,
				NumOfYear:       d.NumOfYear,
				Xd:              d.Xd,
				CreateTimestamp: d.CreateTimestamp,
				ModifyTimestamp: d.ModifyTimestamp,
			}
			result[i] = n
		} else {
			result[i] = d
		}

		if d.Xd != 1.0 {
			xd = d.Xd
		}
	}

	return result, nil
}

func QuoteWithSelectManyLatest(exec mysql.Exec, mode string, code string, date string, limit int64, timeout time.Duration) ([]*model.Quote, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = fmt.Sprintf("select id, code, open, close, high, low, yesterday_closed, volume, account, date, num_of_year, xd, create_timestamp, modify_timestamp from quote_%s where code = ? and DATE_FORMAT(`date`, '%%Y-%%m-%%d') <= ? order by `date` desc limit ?", mode)
	rows, err := exec.QueryContext(ctx, _sql, code, date, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data = make([]*model.Quote, 0, limit)
	for rows.Next() {
		var m = model.Quote{}
		if err := rows.Scan(
			&m.Id,
			&m.Code,
			&m.Open,
			&m.Close,
			&m.High,
			&m.Low,
			&m.YesterdayClosed,
			&m.Volume,
			&m.Account,
			&m.Date,
			&m.NumOfYear,
			&m.Xd,
			&m.CreateTimestamp,
			&m.ModifyTimestamp,
		); err != nil {
			return nil, err
		}
		data = append(data, &m)
	}

	var result = make([]*model.Quote, 0, len(data))
	var xd = 1.0
	for _, d := range data {
		if xd != 1.0 {
			var n = &model.Quote{
				Id:              d.Id,
				Code:            d.Code,
				Open:            mathutil.Trunc2(d.Open * xd),
				Close:           mathutil.Trunc2(d.Close * xd),
				High:            mathutil.Trunc2(d.High * xd),
				Low:             mathutil.Trunc2(d.Low * xd),
				YesterdayClosed: mathutil.Trunc2(d.YesterdayClosed * xd),
				Volume:          d.Volume,
				Account:         d.Account,
				Date:            d.Date,
				NumOfYear:       d.NumOfYear,
				Xd:              d.Xd,
				CreateTimestamp: d.CreateTimestamp,
				ModifyTimestamp: d.ModifyTimestamp,
			}
			result = append(result, n)
		} else {
			result = append(result, d)
		}

		if d.Xd != 1.0 {
			xd = d.Xd
		}
	}

	return result, nil
}

func QuoteWithSelectRangeByDate(exec mysql.Exec, mode string, date string, offset, limit int64, timeout time.Duration) ([]*model.Quote, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = fmt.Sprintf("select id, code, open, close, high, low, yesterday_closed, volume, account, date, num_of_year, xd, create_timestamp, modify_timestamp from quote_%s where DATE_FORMAT(`date`, '%%Y-%%m-%%d') = ? limit ?, ?", mode)
	rows, err := exec.QueryContext(ctx, _sql, date, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data = make([]*model.Quote, 0, limit)
	for rows.Next() {
		var m = model.Quote{}
		if err := rows.Scan(
			&m.Id,
			&m.Code,
			&m.Open,
			&m.Close,
			&m.High,
			&m.Low,
			&m.YesterdayClosed,
			&m.Volume,
			&m.Account,
			&m.Date,
			&m.NumOfYear,
			&m.Xd,
			&m.CreateTimestamp,
			&m.ModifyTimestamp,
		); err != nil {
			return nil, err
		}
		data = append(data, &m)
	}

	return data, nil
}

func QuoteWithSelectOneByCodeAndDate(exec mysql.Exec, mode string, code string, date string, timeout time.Duration) (*model.Quote, error) {
	ctx, cannel := context.WithTimeout(context.Background(), timeout)
	defer cannel()

	var _sql = fmt.Sprintf("select id, code, open, close, high, low, yesterday_closed, volume, account, date, num_of_year, xd, create_timestamp, modify_timestamp from quote_%s where code = ? and DATE_FORMAT(`date`, '%%Y-%%m-%%d') = ?", mode)
	row := exec.QueryRowContext(ctx, _sql, code, date)
	if row.Err() != nil {
		return nil, row.Err()
	}
	var m = model.Quote{}
	if err := row.Scan(
		&m.Id,
		&m.Code,
		&m.Open,
		&m.Close,
		&m.High,
		&m.Low,
		&m.YesterdayClosed,
		&m.Volume,
		&m.Account,
		&m.Date,
		&m.NumOfYear,
		&m.Xd,
		&m.CreateTimestamp,
		&m.ModifyTimestamp,
	); err != nil {
		return nil, err
	}
	return &m, nil
}

const (
	FieldQuoteID              = "id"
	FieldQuoteCode            = "code"
	FieldQuoteOpen            = "open"
	FieldQuoteClose           = "close"
	FieldQuoteHigh            = "high"
	FieldQuoteLow             = "low"
	FieldQuoteYesterdayClosed = "yesterday_closed"
	FieldQuoteVolume          = "volume"
	FieldQuoteAccount         = "account"
	FieldQuoteDate            = "date"
	FieldQuoteNumOfYear       = "num_of_year"
	FieldQuoteXd              = "xd"
	FieldQuoteCreateTimestamp = "create_timestamp"
	FieldQuoteModifyTimestamp = "modify_timestamp"
)

var quoteFeilds = []string{
	FieldQuoteCode,
	FieldQuoteOpen,
	FieldQuoteClose,
	FieldQuoteHigh,
	FieldQuoteLow,
	FieldQuoteYesterdayClosed,
	FieldQuoteVolume,
	FieldQuoteAccount,
	FieldQuoteDate,
	FieldQuoteNumOfYear,
	FieldQuoteXd,
	FieldQuoteCreateTimestamp,
}
