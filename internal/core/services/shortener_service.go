package services

import (
	"URLRotatorGo/infra/logger"
	"URLRotatorGo/infra/workerpool"
	"URLRotatorGo/internal/core/domain"
	"URLRotatorGo/internal/core/ports"
	"URLRotatorGo/pkg"
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ShortenerService struct {
	ShortCodeRepository ports.ShortCodeRepository
	URLRepository       ports.URLRepository
	CacheRepository     ports.CacheRepository
}

func NewShortenerService(ShortCodeRepository ports.ShortCodeRepository, URLRepository ports.URLRepository, CacheRepository ports.CacheRepository) ports.ShortenerService {
	return &ShortenerService{
		ShortCodeRepository: ShortCodeRepository,
		URLRepository:       URLRepository,
		CacheRepository:     CacheRepository,
	}
}

func (s *ShortenerService) GetRedirectURL(ctx context.Context, code string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	shortcode, err := s.CacheRepository.GetShortCode(ctx, code)
	if err != nil {
		logger.L.Info("no cache data for shortcode:", code)

		shortcode, err = s.ShortCodeRepository.GetShortCode(ctx, code)
		if err != nil {
			if errors.Is(err, domain.ErrDataNotFound) {
				return "", domain.ErrDataNotFound
			}
			return "", err
		}

		logger.L.Info("saving data to cache database")
		_ = workerpool.Pool.Submit(func() {
			myctx, mycancel := context.WithTimeout(context.Background(), time.Second*10)
			defer mycancel()
			_ = s.CacheRepository.SaveShortCode(myctx, shortcode)
		})
	}

	links, err := s.CacheRepository.GetLinks(ctx, code)
	if err != nil || len(links) == 0 {
		logger.L.Info("no cache data for links with code:", code)
		links, err = s.URLRepository.GetLinks(ctx, code)
		if err != nil {
			if errors.Is(err, domain.ErrDataNotFound) {
				return "", domain.ErrDataNotFound
			}
			logger.L.Errorw("error while getlinks", "error", err.Error())
			return "", domain.ErrInternalServerError
		}

		_ = workerpool.Pool.Submit(func() {
			myctx, mycancel := context.WithTimeout(context.Background(), time.Second*10)
			defer mycancel()
			_ = s.CacheRepository.SaveLinks(myctx, links)
		})
	}

	if len(links) == 0 {
		return "", domain.ErrDataNotFound
	}

	sort.Slice(links, func(i, j int) bool {
		return links[i].TotalHit < links[j].TotalHit
	})

	var link *domain.URL
	switch shortcode.Strategy {
	case domain.Random:
		link = links[pkg.GenerateRandomNumber(len(links))]
	case domain.RoundRobin:
		link = links[0]
	}

	defer workerpool.Pool.Submit(func() {
		myctx, mycancel := context.WithTimeout(context.Background(), time.Second*10)
		defer mycancel()

		linkID := strconv.Itoa(link.ID)
		_ = s.URLRepository.UpdateHit(myctx, linkID)
		_ = s.ShortCodeRepository.UpdateHit(myctx, shortcode.Code)
		_ = s.CacheRepository.IncrShortCode(myctx, shortcode.Code)
		_ = s.CacheRepository.IncrLink(myctx, code, linkID)
	})

	return link.Original, nil
}

func (s *ShortenerService) ShortURL(ctx context.Context, urls []string, strategy string) (*domain.ShortCode, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	strategy = strings.ToUpper(strategy)
	var strategyAlgo domain.Strategy
	switch strategy {
	case string(domain.Random):
		strategyAlgo = domain.Random
	case string(domain.RoundRobin):
		strategyAlgo = domain.RoundRobin
	default:
		strategyAlgo = domain.RoundRobin
	}

	code := pkg.GenerateShortID()

	shortcode := &domain.ShortCode{
		Code:     code,
		Strategy: strategyAlgo,
	}

	shortcode, err := s.ShortCodeRepository.Save(ctx, shortcode)
	if err != nil {
		return nil, err
	}

	var links []*domain.URL
	for _, url := range urls {
		links = append(links, &domain.URL{
			ShortCode: shortcode.Code,
			Original:  url,
		})
	}

	if links, err = s.URLRepository.Save(ctx, links); err != nil {
		return nil, err
	}

	_ = workerpool.Pool.Submit(func() {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if err = s.CacheRepository.SaveShortCode(ctx, shortcode); err != nil {
			logger.L.Errorw("failed to save cache", "error", err.Error())
		}
		if err = s.CacheRepository.SaveLinks(ctx, links); err != nil {
			logger.L.Errorw("failed to save cache", "error", err.Error())
		}
	})

	return shortcode, nil
}
