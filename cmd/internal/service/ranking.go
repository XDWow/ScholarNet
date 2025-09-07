package service

import (
	"context"
	"errors"
	intrv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/intr/v1"
	"github.com/XD/ScholarNet/cmd/internal/domain"
	"github.com/XD/ScholarNet/cmd/internal/repository"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
)

type RankingService interface {
	TopN(ctx context.Context) error
	//TopN(ctx context.Context, n int64) error
}

type BatchRankingService struct {
	artSvc    ArticleService
	intrSvc   intrv1.InteractiveServiceClient
	repo      repository.RankingRepository
	batchSize int
	n         int
	// scoreFunc 不能返回负数
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService, intrSvc intrv1.InteractiveServiceClient, repo repository.RankingRepository) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		repo:      repo,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(float64(sec+2), 1.5)
		},
	}
}

func (s *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := s.topN(ctx)
	if err != nil {
		return err
	}
	// 存缓存
	return s.repo.ReplaceTopN(ctx, arts)
}

// 返回arts，方便测试
func (s *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	// 只取七天内的数据，超过七天的可以认为绝对不可能成为热榜
	now := time.Now()
	offset := 0

	// 构造基于最小根堆的优先队列
	type Score struct {
		art   domain.Article
		score float64
	}
	// 这里可以用非并发安全
	topN := queue.NewConcurrentPriorityQueue[Score](s.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})

	for {
		// 拿一批
		arts, err := s.artSvc.ListPub(ctx, now, offset, s.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map[domain.Article, int64](arts, func(idx int, src domain.Article) int64 {
			return src.Id
		})
		// 要去找到对应的点赞数据
		resp, err := s.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz:    "article",
			BizIds: ids,
		})
		if err != nil {
			return nil, err
		}
		if len(resp.Intrs) == 0 {
			return nil, errors.New("没有数据")
		}
		// 计算 score
		// 并决定是否要放入优先队列，即topN
		for _, art := range arts {
			intr := resp.Intrs[art.Id]
			score := s.scoreFunc(art.Utime, intr.LikeCnt)
			// 我要考虑，我这个 score 在不在前一百名
			// 两种情况：1、队列未满，直接入 2、队列满了，跟堆顶最小的比
			err = topN.Enqueue(Score{
				art:   art,
				score: score,
			})

			if err == queue.ErrOutOfCapacity {
				val, _ := topN.Peek()
				if val.score < score {
					_, _ = topN.Dequeue()
					err = topN.Enqueue(Score{
						art:   art,
						score: score,
					})
				}
			}
		}
		// 一批已经处理完了，问题来了，我要不要进入下一批？我怎么知道还有没有？
		if len(arts) < s.batchSize || now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			// 我这一批都没取够，我当然没有下一批了
			// 又或者已经取到了七天之前的数据了，说明可以中断了
			break
		}
		// 下一批
		offset += len(arts)
	}
	res := make([]domain.Article, s.n)
	for i := s.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			// 取完了，不够n
			break
		}
		res[i] = val.art
	}
	return res, nil
}
