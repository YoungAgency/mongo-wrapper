package stream

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
)

type OffsetManagerList interface {
	OffsetManager
	SetOffsetAndPush(ctx context.Context, offset StreamOffset, msg []byte) error
	SetOffsetAndPublish(ctx context.Context, offset *StreamOffset, channel, msg string) error
}

type OffsetManager interface {
	GetOffset(ctx context.Context) (*StreamOffset, error)
	SetOffset(ctx context.Context, offset StreamOffset) error
}

type defaultOffsetManager struct {
}

func (d *defaultOffsetManager) GetOffset(ctx context.Context) (*StreamOffset, error) {
	return nil, nil
}

func (d *defaultOffsetManager) SetOffset(ctx context.Context, offset StreamOffset) error {
	return nil
}

func (d *defaultOffsetManager) SetOffsetAndPush(ctx context.Context, offset StreamOffset, msg []byte) error {
	return nil
}

func (d *defaultOffsetManager) SetOffsetAndPublish(ctx context.Context, offset *StreamOffset, channel, msg string) error {
	return nil
}

type RedisOffsetManager struct {
	pool    *redis.Pool
	hash    string
	key     string
	listKey string
}

type RedisOffsetManagerConfig struct {
	Hash    string
	Key     string
	ListKey string
}

func NewRedisTokenManager(pool *redis.Pool, hash, key, listKey string) *RedisOffsetManager {
	if hash == "" {
		panic("hash can't be empty")
	}
	if key == "" {
		panic("key can't be empty")
	}
	return &RedisOffsetManager{
		pool:    pool,
		hash:    hash,
		key:     key,
		listKey: listKey,
	}
}

func (d *RedisOffsetManager) GetOffset(ctx context.Context) (*StreamOffset, error) {
	conn, err := d.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	val, err := redis.Strings(conn.Do("HMGET", d.hash, d.key, d.key+"_ts"))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	if len(val) != 2 {
		return nil, nil
	}
	if val[0] == "" && val[1] == "" {
		return nil, nil
	}
	ts, err := time.Parse(time.RFC3339, val[1])
	if err != nil {
		ts = time.Time{}
	}
	ret := &StreamOffset{
		ResumeToken: val[0],
		Timestamp:   ts,
	}
	return ret, nil
}

func (d *RedisOffsetManager) SetOffset(ctx context.Context, offset StreamOffset) error {
	conn, err := d.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	offset.Timestamp = offset.Timestamp.UTC()
	if _, err := conn.Do("HSET", d.hash,
		d.key, offset.ResumeToken,
		d.key+"_ts", offset.Timestamp.Format(time.RFC3339)); err != nil {
		return err
	}
	return nil
}

func (d *RedisOffsetManager) SetOffsetAndPush(ctx context.Context, offset StreamOffset, msg []byte) error {
	conn, err := d.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	offset.Timestamp = offset.Timestamp.UTC()
	if err := conn.Send("MULTI"); err != nil {
		return err
	}
	if err := conn.Send("HSET", d.hash,
		d.key, offset.ResumeToken,
		d.key+"_ts", offset.Timestamp.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := conn.Send("LPUSH", d.listKey, msg); err != nil {
		return err
	}
	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}
	return nil
}

func (d *RedisOffsetManager) SetOffsetAndPublish(ctx context.Context, offset *StreamOffset, channel string, msg string) error {
	conn, err := d.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	offset.Timestamp = offset.Timestamp.UTC()
	if err := conn.Send("MULTI"); err != nil {
		return err
	}
	if err := conn.Send("HSET", d.hash,
		d.key, offset.ResumeToken,
		d.key+"_ts", offset.Timestamp.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := conn.Send("PUBLISH", channel, msg); err != nil {
		return err
	}
	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}
	return nil
}
