package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"TaskScheduler/internal/entities"
)

type postgresRedisRepository struct {
	db       *sql.DB
	rdb      *redis.Client
	cacheTTL time.Duration
}

// NewPostgresRedisRepository creates a new Repository backed by PostgreSQL and Redis.
func NewPostgresRedisRepository(db *sql.DB, rdb *redis.Client) Repository {
	return &postgresRedisRepository{
		db:       db,
		rdb:      rdb,
		cacheTTL: 10 * time.Minute,
	}
}

func (r *postgresRedisRepository) cacheKey(id string) string {
	return "task:" + id
}

func (r *postgresRedisRepository) Create(task *entities.Task) error {
	query := `INSERT INTO tasks (id, title, description, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, task.ID, task.Title, task.Description, task.Status, task.CreatedAt)
	return err
}

func (r *postgresRedisRepository) Get(id string) (*entities.Task, error) {
	ctx := context.Background()
	key := r.cacheKey(id)

	// 1. Try to get from Redis cache
	val, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var task entities.Task
		if err := json.Unmarshal([]byte(val), &task); err == nil {
			log.Printf("[Repository] Cache HIT: task ID %s retrieved from Redis", id)
			return &task, nil
		}
	}

	log.Printf("[Repository] Cache MISS: task ID %s not found in Redis. Querying PostgreSQL...", id)

	// 2. Cache miss: fetch from PostgreSQL
	var t entities.Task
	query := `SELECT id, title, description, status, created_at FROM tasks WHERE id = $1`
	err = r.db.QueryRow(query, id).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[Repository] Database: task ID %s not found in PostgreSQL", id)
			return nil, ErrTaskNotFound
		}
		log.Printf("[Repository] Database error for task ID %s: %v", id, err)
		return nil, err
	}

	log.Printf("[Repository] Database HIT: task ID %s retrieved from PostgreSQL", id)

	// 3. Write back to Redis cache
	data, err := json.Marshal(t)
	if err == nil {
		if err := r.rdb.Set(ctx, key, data, r.cacheTTL).Err(); err != nil {
			log.Printf("[Repository] Cache write failed for task ID %s: %v", id, err)
		} else {
			log.Printf("[Repository] Cache write success for task ID %s", id)
		}
	}

	return &t, nil
}

func (r *postgresRedisRepository) Update(id string, updateFn func(t *entities.Task)) (*entities.Task, error) {
	ctx := context.Background()

	// Use a database transaction to ensure update consistency and row locking
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Get task from DB and lock the row
	var t entities.Task
	query := `SELECT id, title, description, status, created_at FROM tasks WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	// 2. Mutate state via callback
	updateFn(&t)

	// 3. Persist modifications back to DB
	updateQuery := `UPDATE tasks SET title = $1, description = $2, status = $3 WHERE id = $4`
	_, err = tx.ExecContext(ctx, updateQuery, t.Title, t.Description, t.Status, t.ID)
	if err != nil {
		return nil, err
	}

	// 4. Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 5. Invalidate Redis cache
	key := r.cacheKey(id)
	_ = r.rdb.Del(ctx, key).Err()

	return &t, nil
}
