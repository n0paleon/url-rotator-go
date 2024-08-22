package cache

import (
	"URLRotatorGo/infra/database"
	"URLRotatorGo/infra/logger"
	"URLRotatorGo/internal/core/domain"
	"URLRotatorGo/internal/core/ports"
	"context"
	"strconv"
	"time"

	"github.com/bsm/redislock"
	"github.com/bytedance/sonic"
)

type RedisCache struct {
	db     *database.Redis
	locker *redislock.Client
}

var (
	ShortCodePrefix   = "code:"
	LinksPrefix       = "links:"
	LockPrefix        = "lock:"
	RotatePrefix      = "rotate-id:"
	LockTimeout       = time.Second * 5
	DefaultExpiration = 30 * 24 * time.Hour
)

func NewRedisCache(db *database.Redis) ports.CacheRepository {
	locker := redislock.New(db.Client)
	return &RedisCache{
		db:     db,
		locker: locker,
	}
}

func (r *RedisCache) SaveLinks(ctx context.Context, links []*domain.URL) error {
	pipe := r.db.Client.TxPipeline()

	for _, link := range links {
		id := LinksPrefix + link.ShortCode + ":" + RotatePrefix + strconv.Itoa(link.ID)
		data := map[string]interface{}{
			"id":         link.ID,
			"shortcode":  link.ShortCode,
			"original":   link.Original,
			"total_hit":  link.TotalHit,
			"created_at": link.CreatedAt,
			"updated_at": link.UpdatedAt,
		}

		pipe.HSet(ctx, id, data)
		pipe.Expire(ctx, LinksPrefix+link.ShortCode, DefaultExpiration)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.L.Errorw("failed to save links", "error", err.Error())
		return err
	}

	return nil
}

func (r *RedisCache) IncrLink(ctx context.Context, shortCode, id string) error {
	backoff := redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 5)

	l, err := r.locker.Obtain(ctx, LockPrefix+LinksPrefix+shortCode+":"+RotatePrefix+id, LockTimeout, &redislock.Options{
		RetryStrategy: backoff,
	})
	if err != nil {
		logger.L.Errorw("failed to obtain locker", "shortcode", shortCode, "error", err.Error())
		return err
	}
	defer l.Release(ctx)

	pipe := r.db.TxPipeline()
	target := LinksPrefix + shortCode + ":" + RotatePrefix + id

	pipe.HIncrBy(ctx, target, "total_hit", 1)
	pipe.HSet(ctx, target, "updated_at", time.Now())

	_, err = pipe.Exec(ctx)
	if err != nil {
		logger.L.Errorw("failed to incr link", "shortcode", shortCode, "error", err.Error())
		return err
	}

	return nil
}

func (r *RedisCache) GetLinks(ctx context.Context, code string) ([]*domain.URL, error) {
	var cursor uint64
	var results = make(map[string]map[string]string)
	var finalData []*domain.URL

	for {
		keys, nextCursor, err := r.db.Client.Scan(ctx, cursor, LinksPrefix+code+"*", 0).Result()
		if err != nil {
			logger.L.Errorw("failed to scan keys", "error", err.Error())
			return nil, err
		}
		cursor = nextCursor

		for _, key := range keys {
			data, err := r.db.Client.HGetAll(ctx, key).Result()
			if err != nil {
				logger.L.Errorw("failed to get data for key", "key", key, "error", err.Error())
				return nil, err
			}
			results[key] = data
		}

		if cursor == 0 {
			break
		}
	}

	for _, value := range results {
		var url domain.URL

		// Parse each field from string to the appropriate type
		if idStr, ok := value["id"]; ok {
			url.ID, _ = strconv.Atoi(idStr)
		}
		if totalHitStr, ok := value["total_hit"]; ok {
			url.TotalHit, _ = strconv.Atoi(totalHitStr)
		}
		url.ShortCode = value["shortcode"]
		url.Original = value["original"]
		url.CreatedAt, _ = time.Parse(time.RFC3339, value["created_at"])
		url.UpdatedAt, _ = time.Parse(time.RFC3339, value["updated_at"])

		finalData = append(finalData, &url)
	}

	return finalData, nil
}

func (r *RedisCache) SaveShortCode(ctx context.Context, shortcode *domain.ShortCode) error {
	pipe := r.db.TxPipeline()

	pipe.HSet(ctx, ShortCodePrefix+shortcode.Code, map[string]interface{}{
		"id":         shortcode.ID,
		"code":       shortcode.Code,
		"total_hit":  shortcode.TotalHit,
		"strategy":   string(shortcode.Strategy),
		"created_at": shortcode.CreatedAt,
		"updated_at": shortcode.UpdatedAt,
	})
	pipe.Expire(ctx, ShortCodePrefix+shortcode.Code, DefaultExpiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.L.Errorw("failed to save short code to redis storage", "error", err.Error())
		return err
	}

	return nil
}

func (r *RedisCache) GetShortCode(ctx context.Context, code string) (*domain.ShortCode, error) {
	jsonmap, err := r.db.HGetAll(ctx, ShortCodePrefix+code).Result()
	if err != nil || len(jsonmap) < 1 {
		return nil, domain.ErrDataNotFound
	}

	jsonbyte, err := sonic.Marshal(jsonmap)
	if err != nil {
		logger.L.Errorw("failed to get short code from redis storage", "error", err.Error())
		return nil, err
	}

	var shortcode domain.ShortCode
	if err = sonic.Unmarshal(jsonbyte, &shortcode); err != nil {
		logger.L.Errorw("failed to get short code from redis storage", "error", err.Error())
		return nil, err
	}

	return &shortcode, nil
}

func (r *RedisCache) IncrShortCode(ctx context.Context, code string) error {
	backoff := redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 5)

	l, err := r.locker.Obtain(ctx, LockPrefix+ShortCodePrefix+code, LockTimeout, &redislock.Options{
		RetryStrategy: backoff,
	})
	if err != nil {
		logger.L.Errorw("failed to incr short code in redis storage", "error", err.Error())
		return err
	}
	defer l.Release(ctx)

	pipe := r.db.TxPipeline()
	target := ShortCodePrefix + code
	pipe.HIncrBy(ctx, target, "total_hit", 1)
	pipe.HSet(ctx, target, "updated_at", time.Now())

	_, err = pipe.Exec(ctx)
	if err != nil {
		logger.L.Errorw("failed to incr short code in redis storage", "error", err.Error())
		return err
	}

	return nil
}
