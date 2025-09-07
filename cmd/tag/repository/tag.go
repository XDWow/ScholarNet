package repository

import (
	"context"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"github.com/XD/ScholarNet/cmd/tag/domain"
	"github.com/XD/ScholarNet/cmd/tag/repository/cache"
	"github.com/XD/ScholarNet/cmd/tag/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type TagRepository interface {
	CreateTag(ctx context.Context, tag domain.Tag) (int64, error)
	BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error)
	GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error)
	PreloadUserTags(ctx context.Context) error
}

type CachedTagRepository struct {
	dao   dao.TagDAO
	cache cache.TagCache
	l     logger.LoggerV1
}

func NewTagRepository(dao dao.TagDAO, cache cache.TagCache, l logger.LoggerV1) TagRepository {
	return &CachedTagRepository{
		dao:   dao,
		cache: cache,
		l:     l,
	}
}

// PreloadUserTags 在 toB 的场景下，你可以提前预加载缓存
func (repo *CachedTagRepository) PreloadUserTags(ctx context.Context) error {
	// 我怎么预加载？
	// 缓存里面，究竟怎么存？
	// 1. 放 json，json 里面是一个用户的所有的标签 uid => [{}, {}]
	// 按照用户 ID 来查找
	//var uid int64= 1
	//for {
	//	repo.dao.GetTagsByUid(ctx, uid)
	//	uid ++
	//}
	// select * from tags group by uid
	// 使用 redis 的数据结构
	// 1. list
	// 2. hash 用 hash 结构
	// 3. set, sorted set 都可以

	offset := 0
	batch := 100
	for {
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		// 在这里还有一点点的优化手段，就是 GetTags 的时候，order by uid
		tags, err := repo.dao.GetTags(dbCtx, offset, batch)
		cancel()
		if err != nil {
			// 记录日志，然后返回
			return err
		}

		// 按照 uid 进行分组，一个 uid 执行一次 append

		// 这些 tag 是归属于不同的用户
		for _, tag := range tags {
			rctx, cancel := context.WithTimeout(ctx, time.Second)
			err = repo.cache.Append(rctx, tag.Uid, repo.toDomain(tag))
			cancel()
			if err != nil {
				// 记录日志，你可以中断，你也可以继续
				continue
			}
		}
		if len(tags) < batch {
			return nil
		}
		offset += batch
	}
}

func (repo *CachedTagRepository) CreateTag(ctx context.Context, tag domain.Tag) (int64, error) {
	id, err := repo.dao.CreateTag(ctx, repo.toEntity(tag))
	if err != nil {
		return 0, err
	}
	// 也可以考虑用 DelTags
	// 记得更新你的缓存
	err = repo.cache.Append(ctx, tag.Uid, tag)
	if err != nil {
		repo.l.Warn("创建 Tag 更新缓存失败")
	}
	return id, nil
}

func (repo *CachedTagRepository) BindTagToBiz(ctx context.Context, uid int64,
	biz string, bizId int64, tags []int64) error {
	// 按照我们的说法，我们是要覆盖式地执行打标签
	// 新的标签完全覆盖老的标签
	err := repo.dao.DeleteTagBiz(ctx, uid, biz, bizId)
	if err != nil {
		return err
	}
	return repo.dao.CreateTagBiz(ctx, dao.TagBiz{
		BizId: bizId,
		Biz:   biz,
		Uid:   uid,
		Tid:   tags,
	})
}

func (repo *CachedTagRepository) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	res, err := repo.cache.GetTags(ctx, uid)
	if err == nil {
		return res, err
	}
	// 下面也是慢路径，你同样可以说降级的时候不执行

	// 如果我要缓存
	// 我这里应该是 uid => tags 的映射
	// toB 的时候，我直接全量缓存
	// 我要在应用启动的时候，把缓存加载好
	// 如果你认为你的 tags 是没有过期时间的，你这里就不用回查数据库了
	tags, err := repo.dao.GetTagsByUid(ctx, uid)
	if err != nil {
		return nil, err
	}

	res = slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	})
	// 记得回写缓存
	err = repo.cache.Append(ctx, uid, res...)
	if err != nil {
		// 记录日志就行，缓存回写失败，不认为是一个问题
		repo.l.Warn("Tag 回写缓存失败")
	}
	return res, err
}

func (repo *CachedTagRepository) GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error) {
	tags, err := repo.dao.GetTagsById(ctx, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	}), nil
}

func (repo *CachedTagRepository) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	// 如果要缓存的话，就是 uid + biz + biz_id 构成一个 key
	tags, err := repo.dao.GetTagsByBiz(ctx, uid, biz, bizId)
	if err != nil {
		return nil, err
	}
	return slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	}), nil
}

func (repo *CachedTagRepository) toEntity(tag domain.Tag) dao.Tag {
	return dao.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}

func (repo *CachedTagRepository) toDomain(tag dao.Tag) domain.Tag {
	return domain.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}
