package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/article/domain"
	"github.com/XD/ScholarNet/cmd/article/repository/cache"
	"github.com/XD/ScholarNet/cmd/article/repository/dao"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

//go:generate mockgen -source=./type.go -package=repomocks -destination=mocks/article.mock.go ArticleRepository
type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	List(ctx context.Context, author int64,
		offset int, limit int) ([]domain.Article, error)

	// Sync 本身要求先保存到制作库，再同步到线上库
	Sync(ctx context.Context, art domain.Article) (int64, error)
	// SyncStatus 仅仅同步状态
	SyncStatus(ctx context.Context, uid, id int64, status domain.ArticleStatus) error
	GetById(ctx context.Context, id int64) (domain.Article, error)

	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, utime time.Time, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	// 操作单一的库
	dao   dao.ArticleDAO
	cache cache.ArticleCache

	// SyncV1 用
	authorDAO dao.ArticleAuthorDAO
	readerDAO dao.ArticleReaderDAO

	// SyncV2 用
	db *gorm.DB
	l  logger.LoggerV1
}

func (repo *CachedArticleRepository) Cache() cache.ArticleCache {
	return repo.cache
}

func (repo *CachedArticleRepository) ListPub(ctx context.Context, utime time.Time, offset int, limit int) ([]domain.Article, error) {
	val, err := repo.dao.ListPubByUtime(ctx, utime, offset, limit)
	if err != nil {
		return nil, err
	}
	res, err := slice.Map[dao.PublishedArticle, domain.Article](val, func(idx int, src dao.PublishedArticle) domain.Article {
		return repo.ToDomain(dao.Article(src))
	}), nil

	go func() {
		repo.prePublishCache(ctx, res)
	}()

	return res, err
}

func NewArticleRepository(dao dao.ArticleDAO,
	c cache.ArticleCache,
	l logger.LoggerV1) ArticleRepository {
	return &CachedArticleRepository{
		dao:   dao,
		l:     l,
		cache: c,
	}
}

func NewArticleRepositoryV1(authorDAO dao.ArticleAuthorDAO,
	readerDAO dao.ArticleReaderDAO) ArticleRepository {
	return &CachedArticleRepository{
		authorDAO: authorDAO,
		readerDAO: readerDAO,
	}
}

func (repo *CachedArticleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := repo.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	}
	art, err := repo.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
	}
	// 也可以同步
	go func() {
		if err = repo.cache.SetPub(ctx, res); err != nil {
			repo.l.Error("缓存已发表文章失败",
				logger.Error(err), logger.Int64("aid", res.Id))
		}
	}()
	return res, nil
}

func (repo *CachedArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	cachedArt, err := repo.cache.Get(ctx, id)
	if err == nil {
		return cachedArt, nil
	}
	art, err := repo.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return repo.ToDomain(art), nil
}

func (repo *CachedArticleRepository) List(ctx context.Context, author int64,
	offset int, limit int) ([]domain.Article, error) {
	// 只有第一页才走缓存，并且假定一页只有 100 条
	// 也就是说，如果前端允许创作者调整页的大小
	// 那么只有 100 这个页大小这个默认情况下，会走索引
	if offset == 0 && limit == 100 {
		data, err := repo.cache.GetFirstPage(ctx, author)
		if err == nil {
			go func() {
				repo.preCache(ctx, data)
			}()
			return data, nil
		}
		// 这里记录日志
		if err != cache.ErrKeyNotExist {
			repo.l.Error("查询缓存文章失败",
				logger.Int64("author", author), logger.Error(err))
		}
	}
	// 慢路径
	arts, err := repo.dao.GetByAuthor(ctx, author, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](arts,
		func(idx int, src dao.Article) domain.Article {
			return repo.ToDomain(src)
		})
	// 一般都是让调用者来控制是否异步。
	go func() {
		repo.preCache(ctx, res)
	}()
	// 你这个也可以做成异步的
	err = repo.cache.SetFirstPage(ctx, author, res)
	if err != nil {
		repo.l.Error("刷新第一页文章的缓存失败",
			logger.Int64("author", author), logger.Error(err))
	}
	return res, nil
}

func (repo *CachedArticleRepository) preCache(ctx context.Context,
	arts []domain.Article) {
	// 1MB
	const contentSizeThreshold = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) <= contentSizeThreshold {
		// 你也可以记录日志
		if err := repo.cache.Set(ctx, arts[0]); err != nil {
			repo.l.Error("提前准备缓存失败", logger.Error(err))
		}
	}
}

func (repo *CachedArticleRepository) prePublishCache(ctx context.Context,
	arts []domain.Article) {
	// 1MB
	const contentSizeThreshold = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) <= contentSizeThreshold {
		// 你也可以记录日志
		if err := repo.cache.PreSetPub(ctx, arts[0]); err != nil {
			repo.l.Error("提前准备缓存失败", logger.Error(err))
		}
	}
}

func (repo *CachedArticleRepository) SyncStatus(ctx context.Context,
	uid, id int64, status domain.ArticleStatus) error {
	return repo.dao.SyncStatus(ctx, uid, id, status.ToUint8())
}

func (repo *CachedArticleRepository) Sync(ctx context.Context,
	art domain.Article) (int64, error) {
	id, err := repo.dao.Sync(ctx, repo.toEntity(art))
	if err != nil {
		return 0, err
	}
	go func() {
		author := art.Author.Id
		err = repo.cache.DelFirstPage(ctx, author)
		if err != nil {
			repo.l.Error("删除第一页缓存失败",
				logger.Int64("author", author), logger.Error(err))
		}
		err = repo.cache.SetPub(ctx, art)
		if err != nil {
			repo.l.Error("提前设置缓存失败",
				logger.Int64("author", author), logger.Error(err))
		}
	}()
	return id, nil
}

func (repo *CachedArticleRepository) SyncV2(ctx context.Context,
	art domain.Article) (int64, error) {
	tx := repo.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 直接 defer Rollback
	// 如果我们后续 Commit 了，这里会得到一个错误，但是没关系
	defer tx.Rollback()
	authorDAO := dao.NewGORMArticleDAO(tx)
	readerDAO := dao.NewGORMArticleReaderDAO(tx)

	// 下面代码和 SyncV1 一模一样
	artn := repo.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = authorDAO.Insert(ctx, artn)
		if err != nil {
			return 0, err
		}
	} else {
		err = authorDAO.UpdateById(ctx, artn)
	}
	if err != nil {
		return 0, err
	}
	artn.Id = id
	err = readerDAO.UpsertV2(ctx, dao.PublishedArticle(artn))
	if err != nil {
		// 依赖于 defer 来 rollback
		return 0, err
	}
	tx.Commit()
	return artn.Id, nil
}

func (repo *CachedArticleRepository) SyncV1(ctx context.Context,
	art domain.Article) (int64, error) {
	artn := repo.toEntity(art)
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = repo.authorDAO.Create(ctx, artn)
		if err != nil {
			return 0, err
		}
	} else {
		err = repo.authorDAO.UpdateById(ctx, artn)
	}
	if err != nil {
		return 0, err
	}
	artn.Id = id
	err = repo.readerDAO.Upsert(ctx, artn)
	return id, err
}

func (repo *CachedArticleRepository) Create(ctx context.Context,
	art domain.Article) (int64, error) {
	id, err := repo.dao.Insert(ctx, repo.toEntity(art))
	if err != nil {
		return 0, err
	}
	author := art.Author.Id
	err = repo.cache.DelFirstPage(ctx, author)
	if err != nil {
		repo.l.Error("删除缓存失败",
			logger.Int64("author", author), logger.Error(err))
	}
	return id, nil
}

func (repo *CachedArticleRepository) Update(ctx context.Context,
	art domain.Article) error {
	err := repo.dao.UpdateById(ctx, repo.toEntity(art))
	if err != nil {
		return err
	}
	author := art.Author.Id
	err = repo.cache.DelFirstPage(ctx, author)
	if err != nil {
		repo.l.Error("删除缓存失败",
			logger.Int64("author", author), logger.Error(err))
	}
	return nil
}

func (repo *CachedArticleRepository) ToDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
	}
}

func (repo *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		// 这一步，就是将领域状态转化为存储状态。
		// 这里我们就是直接转换，
		// 有些情况下，这里可能是借助一个 map 来转
		Status: uint8(art.Status),
	}
}
