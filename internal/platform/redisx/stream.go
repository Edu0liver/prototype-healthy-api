package redisx

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Enqueue publishes a job onto a Redis Stream (durable buffer between ingestion
// and the orchestration workers — provides back-pressure and at-least-once).
func (c *Client) Enqueue(ctx context.Context, stream string, values map[string]any) (string, error) {
	return c.XAdd(ctx, &redis.XAddArgs{Stream: stream, Values: values}).Result()
}

// EnsureGroup creates the consumer group at the stream tail if it does not exist.
func (c *Client) EnsureGroup(ctx context.Context, stream, group string) error {
	err := c.XGroupCreateMkStream(ctx, stream, group, "$").Err()
	if err != nil && !isBusyGroup(err) {
		return err
	}
	return nil
}

// ReadGroup blocks up to `block` for new entries for the given consumer.
func (c *Client) ReadGroup(ctx context.Context, stream, group, consumer string, count int64, block time.Duration) ([]redis.XStream, error) {
	return c.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    block,
	}).Result()
}

// Ack acknowledges processing of a stream entry.
func (c *Client) Ack(ctx context.Context, stream, group, id string) error {
	return c.XAck(ctx, stream, group, id).Err()
}

// isBusyGroup reports whether the error is the harmless "group already exists".
func isBusyGroup(err error) bool {
	return err != nil && !errors.Is(err, redis.Nil) && strings.Contains(err.Error(), "BUSYGROUP")
}
