package article

import (
	"context"
	"github.com/LXD-c/basic-go/webook/internal/domain"
	"github.com/LXD-c/basic-go/webook/internal/repository"
	"github.com/LXD-c/basic-go/webook/internal/repository/cache"
	dao "github.com/LXD-c/basic-go/webook/internal/repository/dao/article"
	"github.com/LXD-c/basic-go/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

// repository 还是要用来操作缓存和DAO
// 事务概念应该在 DAO 这一层

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// Sync 存储并同步数据
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error
	//FindById(ctx context.Context, id int64) domain.Article
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetByID(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedByID(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao      dao.ArticleDAO
	userRepo repository.UserRepository

	// v1 操作两个 DAO
	readerDAO dao.ReaderDAO
	authorDAO dao.AuthorDAO

	// 耦合了 DAO 操作的东西,跨层耦合，不好
	// 正常情况下，如果你要在 repository 层面上操作事务
	// 那么就只能利用 db 开始事务之后，创建基于事务的 DAO
	// 或者，直接去掉 DAO 这一层，在 repository 的实现中，直接操作 db
	db *gorm.DB

	cache cache.ArticleCache
	l     logger.LoggerV1
}

func NewArticleRepository(dao dao.ArticleDAO,
	cache cache.ArticleCache,
	userRepo repository.UserRepository,
	l logger.LoggerV1) ArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		cache:    cache,
		l:        l,
		userRepo: userRepo,
	}
}

func (repo *CachedArticleRepository) ListPub(ctx context.Context,
	start time.Time,
	offset int, limit int) ([]domain.Article, error) {
	res, err := repo.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return repo.toDomain(src)
	}), nil
}

func (repo *CachedArticleRepository) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	data, err := repo.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return repo.toDomain(data), nil
}

func (repo *CachedArticleRepository) GetPublishedByID(ctx context.Context, id int64) (domain.Article, error) {
	// 读取线上库数据，如果你的 Content 被你放过去了 OSS 上，你就要让前端去读 Content 字段
	art, err := repo.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 读者看文章，需要 Author 信息
	// 你在这边要组装 user 了，适合单体应用
	usr, err := repo.userRepo.FindById(ctx, art.AuthorId)
	res := domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id:   usr.Id,
			Name: usr.Nickname,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
	return res, nil
}

func (repo *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 你在这个地方，集成你的复杂的缓存方案
	// 这里是：先去缓存拿，没拿到再去数据库，并且拿到列表后大概率会访问第一条，所以预缓存第一条数据
	// 只缓存第一页
	if offset == 0 && limit < 100 {
		data, err := repo.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				repo.preCache(ctx, data)
			}()
			//return data[:limit], nil
			return data, nil
		}
	}
	res, err := repo.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return repo.toDomain(src)
	})
	// 回写缓存的时候，可以同步，也可以异步
	go func() {
		err := repo.cache.SetFirstPage(ctx, uid, data)
		if err != nil {
			repo.l.Error("回写缓存失败", logger.Error(err))
		}
		repo.preCache(ctx, data)
	}()
	return data, nil
}

func (repo *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error {
	return repo.dao.SyncStatus(ctx, id, author, status.ToUint8())
}

func (repo *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	// 刚刚发布的文章，访问量应该较大，放入缓存
	// 当然是在 repository 层操作 Cache
	id, err := repo.dao.Sync(ctx, repo.toEntity(art))
	if err == nil {
		err = repo.cache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			// 不需要特别关心
			// 比如说输出 WARN 日志
		}
		err = repo.cache.SetPub(ctx, art)
		if err != nil {
			// 不需要特别关心
			// 比如说输出 WARN 日志
		}
	}
	return id, err
}

//func (c *CachedArticleRepository) SyncV2_1(ctx context.Context, art domain.Article) (int64, error) {
//	// 谁在控制事务，是 repository，还是DAO在控制事务？
//	c.dao.Transaction(ctx, func(txDAO dao.ArticleDAO) error {
//
//	})
//}

// SyncV2 尝试在 repository 层面上解决事务问题
// 确保保存到制作库和线上库同时成功，或者同时失败
func (repo *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	// 开启了一个事务
	tx := repo.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 防止中途 panic/return 没有回滚
	defer tx.Rollback()
	// 利用 tx 来构建 DAO
	author := dao.NewAuthorDAO(tx)
	reader := dao.NewReaderDAO(tx)

	var (
		id  = art.Id
		err error
	)
	artn := repo.toEntity(art)
	// 应该先保存到制作库，再保存到线上库
	if id > 0 {
		err = author.UpdateById(ctx, artn)
	} else {
		id, err = author.Insert(ctx, artn)
	}
	if err != nil {
		// 执行有问题，要回滚
		//tx.Rollback()
		return id, err
	}
	// 操作线上库了，保存数据，同步过来
	// 考虑到，此时线上库可能有，可能没有，你要有一个 UPSERT 的写法
	// INSERT or UPDATE
	// 如果数据库有，那么就更新，不然就插入
	err = reader.UpsertV2(ctx, dao.PublishedArticle(artn))
	// 执行成功，直接提交
	tx.Commit()
	return id, err
}

func (repo *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	artn := repo.toEntity(art)
	// 应该先保存到制作库，再保存到线上库
	if id > 0 {
		err = repo.authorDAO.UpdateById(ctx, artn)
	} else {
		id, err = repo.authorDAO.Insert(ctx, artn)
	}
	if err != nil {
		return id, err
	}
	// 操作线上库了，保存数据，同步过来
	// 考虑到，此时线上库可能有，可能没有，你要有一个 UPSERT 的写法
	// INSERT or UPDATE
	// 如果数据库有，那么就更新，不然就插入
	err = repo.readerDAO.Upsert(ctx, artn)
	return id, err
}

func (repo *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return repo.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}

func (repo *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return repo.dao.UpdateById(ctx, dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
	})
}

func (repo *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (repo *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Status: domain.ArticleStatus(art.Status),
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
	}
}

func (repo *CachedArticleRepository) preCache(ctx context.Context, data []domain.Article) {
	// 列表中至少有一篇文章，并且要缓存的第一篇文章不要太大，太大就不保存了，舍弃一点性能，节省空间
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := repo.cache.Set(ctx, data[0])
		if err != nil {
			repo.l.Error("提前预加载缓存失败", logger.Error(err))
		}
	}
}
