package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	DSN     string
	MaxOpen uint64 = 10
	DB      *mongo.Client

	RetryTimes = 3
	Period     = 10 * time.Second
)

var (
	DefaultConnectTimeout = 10 * time.Second
)

func Build() error {
	var (
		pool *mongo.Client
		err  error

		i = 1
	)
	for {
		if i > RetryTimes {
			if err != nil {
				return err
			}
			return fmt.Errorf("panic: connect mongodb failure, err is nil?")
		}
		pool, err = buildMongoDB(DSN)
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
	return err
}

func Close() error {
	if DB == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultConnectTimeout)
	defer cancel()

	return DB.Disconnect(ctx)
}

func buildMongoDB(dsn string) (*mongo.Client, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN no set")
	}

	client, err := mongo.NewClient(
		options.Client().ApplyURI(dsn).SetMaxPoolSize(MaxOpen),
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultConnectTimeout)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}
