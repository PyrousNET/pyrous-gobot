package pubsub

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"net/url"
)

type (
	Pubsub interface {
		Publish(channel string, payload []byte) error
		Subscribe(channel string) *redis.PubSub
		Get(key string) *redis.StringCmd
		Del(key string) *redis.IntCmd
		Close() error
	}

	RedisPubsub struct {
		conn *redis.Client
		ctx  context.Context
	}
)

func GetPubsub(connStr string) Pubsub {
	uri, _ := url.Parse(connStr)

	switch uri.Scheme {
	case "redis":
		return GetRedisPubsub(connStr)
	default:
		return nil
	}
}

func GetRedisPubsub(connStr string) *RedisPubsub {
	uri, _ := url.Parse(connStr)
	password, _ := uri.User.Password()

	rps := &RedisPubsub{
		conn: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", uri.Hostname(), uri.Port()),
			Username: uri.User.Username(),
			Password: password,
		}),
	}

	rps.ctx = context.Background()

	return rps
}

func (rp *RedisPubsub) Publish(channel string, payload []byte) error {
	return rp.conn.Publish(rp.ctx, channel, payload).Err()
}

func (rp *RedisPubsub) Subscribe(channel string) *redis.PubSub {
	return rp.conn.Subscribe(rp.ctx, channel)
}

func (rp *RedisPubsub) Close() error {
	return rp.conn.Close()
}

func (rp *RedisPubsub) Get(key string) *redis.StringCmd {
	return rp.conn.Get(rp.ctx, key)
}

func (rp *RedisPubsub) Del(key string) *redis.IntCmd {
	return rp.conn.Del(rp.ctx, key)
}
