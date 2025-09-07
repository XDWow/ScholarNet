package service

import (
	"context"
	searchv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/search/v1"
	"github.com/XD/ScholarNet/cmd/article/domain"
	"github.com/XD/ScholarNet/cmd/article/events"
	"github.com/XD/ScholarNet/cmd/article/repository"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"golang.org/x/sync/errgroup"
	"time"
)

//go:generate mockgen -source=./type.go -package=svcmocks -destination=mocks/article.mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid, id int64) error
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	List(ctx context.Context, author int64,
		offset, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)

	// 剩下的这个是给读者用的服务，暂时放到这里

	// GetPublishedById 查找已经发表的
	// 正常来说在微服务架构下，读者服务和创作者服务会是两个独立的服务
	// 单体应用下可以混在一起，毕竟现在也没几个方法
	GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error)
	// ListPub 根据更新时间来分页，更新时间必须小于 startTime
	ListPub(ctx context.Context, startTime time.Time, offset, limit int) ([]domain.Article, error)
}

type articleService struct {
	// 1. 在 service 这一层使用两个 repository
	authorRepo repository.ArticleAuthorRepository
	readerRepo repository.ArticleReaderRepository

	// 2. 在 repo 里面处理制作库和线上库
	// 1 和 2 是互斥的，不会同时存在
	userRepo repository.AuthorRepository
	repo     repository.ArticleRepository
	logger   logger.LoggerV1

	syncClient searchv1.SyncServiceClient

	// 搞个异步的
	producer events.Producer
}

func (svc *articleService) ListPub(ctx context.Context,
	startTime time.Time,
	offset, limit int) ([]domain.Article, error) {
	return svc.repo.ListPub(ctx, startTime, offset, limit)
}

func (svc *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return svc.repo.GetById(ctx, id)
}

func (svc *articleService) List(ctx context.Context, author int64,
	offset, limit int) ([]domain.Article, error) {
	return svc.repo.List(ctx, author, offset, limit)
}

func NewArticleService(repo repository.ArticleRepository,
	authorRepo repository.AuthorRepository,
	l logger.LoggerV1,
	producer events.Producer,
) ArticleService {
	return &articleService{
		repo:     repo,
		logger:   l,
		userRepo: authorRepo,
		producer: producer,
	}
}

func NewArticleServiceV1(
	authorRepo repository.ArticleAuthorRepository,
	readerRepo repository.ArticleReaderRepository,
	syncClient searchv1.SyncServiceClient,
	l logger.LoggerV1) ArticleService {
	return &articleService{
		authorRepo: authorRepo,
		readerRepo: readerRepo,
		logger:     l,
		syncClient: syncClient,
	}
}

func (svc *articleService) GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error) {
	var eg errgroup.Group
	var art *domain.Article
	var author *domain.Author
	var err error
	eg.Go(func() error {
		res, eerr := svc.repo.GetPublishedById(ctx, id)
		art = &res
		return eerr
	})
	eg.Go(func() error {
		res, eerr := svc.userRepo.FindAuthor(ctx, id)
		author = &res
		return eerr
	})
	if err = eg.Wait(); err != nil {
		return domain.Article{}, err
	}
	art.Author = *author
	res := *art
	go func() {
		if err == nil {
			er := svc.producer.ProduceReadEvent(events.ReadEvent{
				Aid: id,
				Uid: id,
			})
			if er != nil {
				svc.logger.Error("发送消息失败",
					logger.Int64("uid", uid),
					logger.Int64("aid", id),
					logger.Error(err))
			}
		}
	}()
	return res, err
}

func (svc *articleService) Withdraw(ctx context.Context, uid, id int64) error {
	return svc.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func (svc *articleService) Save(ctx context.Context,
	art domain.Article) (int64, error) {
	// 设置为未发表
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := svc.update(ctx, art)
		return art.Id, err
	}
	return svc.create(ctx, art)
}

func (svc *articleService) Publish(ctx context.Context,
	art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return svc.repo.Sync(ctx, art)
}

// PublishV1 基于使用两种 repository 的写法
func (svc *articleService) PublishV1(ctx context.Context,
	art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	// 这一段逻辑其实就是 Save
	if art.Id == 0 {
		id, err = svc.authorRepo.Create(ctx, art)
	} else {
		err = svc.authorRepo.Update(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	// 保持制作库和线上库的 ID 是一样的。
	art.Id = id
	for i := 0; i < 3; i++ {
		err = svc.readerRepo.Save(ctx, art)
		if err == nil {
			break
		}
		svc.logger.Error("部分失败：保存数据到线上库失败",
			logger.Field{Key: "art_id", Value: id},
			logger.Error(err))
		// 在接入了 metrics 或者 tracing 之后，
		// 这边要进一步记录必要的DEBUG信息。
	}
	if err != nil {
		svc.logger.Error("部分失败：保存数据到线上库重试都失败了",
			logger.Field{Key: "art_id", Value: id},
			logger.Error(err))
		return 0, err
	}
	return id, nil
}

func (svc *articleService) create(ctx context.Context,
	art domain.Article) (int64, error) {
	return svc.repo.Create(ctx, art)
}
func (svc *articleService) update(ctx context.Context,
	art domain.Article) error {
	return svc.repo.Update(ctx, art)
}
