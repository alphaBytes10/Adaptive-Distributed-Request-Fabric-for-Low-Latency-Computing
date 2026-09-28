package cache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	_ "github.com/lib/pq"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

type MultiLevelCache struct {
	l1 sync.Map
	l2 *redis.Client
	l3 *sql.DB
}

func NewMultiLevelCache(redisAddr, pgConnStr string) (*MultiLevelCache, error) {
	// Initialize L2 (Redis)
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Initialize L3 (Postgres)
	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Create table for L3 if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cache_store (
		key VARCHAR(255) PRIMARY KEY,
		value TEXT,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return nil, fmt.Errorf("failed to create l3 table: %w", err)
	}

	return &MultiLevelCache{
		l2: rdb,
		l3: db,
	}, nil
}

func (c *MultiLevelCache) Get(ctx context.Context, key string) (string, string, error) {
	// Try L1
	if val, ok := c.l1.Load(key); ok {
		return val.(string), "L1", nil
	}

	// Try L2
	val, err := c.l2.Get(ctx, key).Result()
	if err == nil {
		// Populate L1
		c.l1.Store(key, val)
		return val, "L2", nil
	} else if err != redis.Nil {
		fmt.Printf("L2 error: %v\n", err)
	}

	// Try L3
	var dbVal string
	err = c.l3.QueryRowContext(ctx, "SELECT value FROM cache_store WHERE key = $1", key).Scan(&dbVal)
	if err == nil {
		// Populate L1 & L2
		c.l1.Store(key, dbVal)
		c.l2.Set(ctx, key, dbVal, 5*time.Minute)
		return dbVal, "L3", nil
	} else if err != sql.ErrNoRows {
		fmt.Printf("L3 error: %v\n", err)
	}

	return "", "", ErrCacheMiss
}

func (c *MultiLevelCache) Set(ctx context.Context, key string, value string) error {
	// Set L1
	c.l1.Store(key, value)

	// Set L2
	c.l2.Set(ctx, key, value, 5*time.Minute)

	// Set L3
	_, err := c.l3.ExecContext(ctx, `
		INSERT INTO cache_store (key, value, updated_at) 
		VALUES ($1, $2, CURRENT_TIMESTAMP) 
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = CURRENT_TIMESTAMP`, 
		key, value)
	return err
}
