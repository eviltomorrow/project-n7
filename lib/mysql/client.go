package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	DSN     string
	MinOpen int = 5
	MaxOpen int = 10
	DB      *sql.DB

	RetryTimes = 3
	Period     = 10 * time.Second
)

var DefaultConnectTimeout = 5 * time.Second

// Build build mysql
func Build() error {
	var (
		pool *sql.DB
		err  error

		i = 1
	)
	for {
		if i > RetryTimes {
			if err != nil {
				return err
			}
			return fmt.Errorf("panic: connect mysql failure, err is nil?")
		}
		pool, err = buildMySQL(DSN)
		if err == nil {
			break
		}
		if err != nil {
			log.Printf("[W] Try to connect to MongoDB=>[Retry: %d], nest error: %v\r\n", i, err)
		}
		i++
		time.Sleep(Period)
	}
	DB = pool

	return nil
}

// Close close mysql
func Close() error {
	if DB == nil {
		return nil
	}

	return DB.Close()
}

func buildMySQL(dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN no set")
	}
	pool, err := sql.Open("mysql", DSN)
	if err != nil {
		return nil, err
	}
	pool.SetConnMaxLifetime(time.Minute * 3)
	pool.SetMaxOpenConns(MaxOpen)
	pool.SetMaxIdleConns(MinOpen)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultConnectTimeout)
	defer cancel()

	if err = pool.PingContext(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}

// Exec exec mysql
type Exec interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
