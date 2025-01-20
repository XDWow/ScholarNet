package service

import (
	"context"
	"github.com/LXD-c/basic-go/webook/internal/domain"
	"github.com/LXD-c/basic-go/webook/internal/repository/article"
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	"time"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
}

// 在哪层区分制作库和线上库，就意味着事务概念在哪层
// 事务概念，一个事务要么全部完成，要么一点都不做
// 这里代表 Publish 后，两个库的同步，制作库修改了，线上库也一定要修改，不然制作库也别动，可以重新 Publish
type ArticleServiceImpl struct {
	repo article.ArticleRepository
	l    logger.LoggerV1

	// V1：在 Service 层上区分制作库、线上库，依靠两个不同的 repository 来解决这种跨表，或者跨库的问题,
	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository
}

func NewArticleServiceV1(author article.ArticleAuthorRepository,
	reader article.ArticleReaderRepository,
	l logger.LoggerV1) ArticleService {
	return &ArticleServiceImpl{
		author: author,
		reader: reader,
		l:      l,
	}
}

func NewArticleService(repo article.ArticleRepository, l logger.LoggerV1) ArticleService {
	return &ArticleServiceImpl{
		repo: repo,
		l:    l,
	}
}

func (s *ArticleServiceImpl) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	// 制作库
	//id, err := a.repo.Create(ctx, art)
	//// 线上库呢？
	//a.repo.SyncToLiveDB(ctx, art)
	return s.repo.Sync(ctx, art)
}

func (s *ArticleServiceImpl) PublishV1(ctx context.Context, article domain.Article) (int64, error) {
	// 发表到制作库,需要判断是更新还是创造
	var (
		err error
		id  = article.Id
	)
	if article.Id > 0 {
		err = s.author.Update(ctx, article)
	} else {
		id, err = s.author.Create(ctx, article)
	}
	if err != nil {
		return 0, err
	}
	article.Id = id
	// 再保存到线上库
	//article.Id, err = s.reader.Save(ctx, article)
	//if err != nil { return 0, err }
	//return article.Id, nil

	// 保存到线上库:重试机制
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second * time.Duration(i))
		id, err = s.reader.Save(ctx, article)
		if err == nil {
			return id, nil
		}
		s.l.Error("帖子保存到制作库成功，保存到线上库失败",
			logger.Int64("article_id", article.Id),
			logger.Error(err),
		)
	}
	s.l.Error("保存到线上库重试彻底失败",
		logger.Int64("article_id", article.Id),
		logger.Error(err))
	return id, err
}

func (s *ArticleServiceImpl) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		return art.Id, s.author.Update(ctx, art)
	}
	return s.author.Create(ctx, art)
}

// 业务层跟数据层的命名风格是不一样的
func (s *ArticleServiceImpl) Withdraw(ctx context.Context, art domain.Article) error {
	// art.Status = domain.ArticleStatusPrivate 然后直接把整个 art 往下传
	return s.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}
