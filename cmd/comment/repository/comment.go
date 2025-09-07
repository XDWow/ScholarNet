package repository

import (
	"context"
	"database/sql"
	"github.com/XD/ScholarNet/cmd/comment/domain"
	"github.com/XD/ScholarNet/cmd/comment/repository/dao"
	"github.com/XD/ScholarNet/cmd/pkg/logger"
	"golang.org/x/sync/errgroup"
	"time"
)

type CommentRepository interface {
	// FindByBiz 根据 ID 倒序查找
	// 并且会返回每个评论的三条直接回复
	FindByBiz(ctx context.Context, biz string,
		bizId, minID, limit int64) ([]domain.Comment, error)
	// DeleteComment 删除评论，删除本评论何其子评论
	DeleteComment(ctx context.Context, comment domain.Comment) error
	// CreateComment 创建评论
	CreateComment(ctx context.Context, comment domain.Comment) error
	// GetCommentByIds 获取单条评论 支持批量获取
	GetCommentByIds(ctx context.Context, id []int64) ([]domain.Comment, error)
	GetMoreReplies(ctx context.Context, rid int64, maxID int64, limit int64) ([]domain.Comment, error)
}

type CachedCommentRepo struct {
	dao dao.CommentDAO
	l   logger.LoggerV1
}

func (c *CachedCommentRepo) GetMoreReplies(ctx context.Context, rid int64, maxID int64, limit int64) ([]domain.Comment, error) {
	cs, err := c.dao.FindRepliesByRid(ctx, rid, maxID, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(cs))
	for _, cm := range cs {
		res = append(res, c.toDomain(cm))
	}
	return res, nil
}

func (c *CachedCommentRepo) FindByBiz(ctx context.Context, biz string,
	bizId, minID, limit int64) ([]domain.Comment, error) {
	// 事实上，最新评论它的缓存效果不是很好
	// 在这里缓存第一页，缓存咩有，就去找数据库
	// 也可以考虑定时刷新缓存
	// 拿到的就是顶级评论
	daoComments, err := c.dao.FindByBiz(ctx, biz, bizId, minID, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(daoComments))
	// 拿到前三条子评论:
	// 按照 pid 来分组，取组内三条（这三条是按照 ID 降序排序）
	// SELECT * FROM `comments` WHERE pid IN $ids GROUP BY pid ORDER BY id DESC LIMIT 3;
	var eg errgroup.Group
	for _, dc := range daoComments {
		dc := dc
		current := c.toDomain(dc)
		res = append(res, current)
		// 降级不需要去查询子评论
		if ctx.Value("downgrade") == "true" {
			// 尤其要关注数据库的读压力
			continue
		}
		eg.Go(func() error {
			// 去数据库查询
			// 取三条回复
			subCs, err := c.dao.FindRepliesByPID(ctx, dc.ID, 0, 3)
			if err != nil {
				return err
			}
			current.Children = make([]domain.Comment, 0, len(subCs))
			// 不然呢？
			for _, sc := range subCs {
				// 构建子评论
				current.Children = append(current.Children, c.toDomain(sc))
			}
			return nil
		})
	}
	return res, eg.Wait()
}

func (c *CachedCommentRepo) DeleteComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Delete(ctx, dao.Comment{
		ID: comment.Id,
	})
}

func (c *CachedCommentRepo) CreateComment(ctx context.Context, comment domain.Comment) error {
	return c.dao.Insert(ctx, c.toEntity(comment))
}

func (c *CachedCommentRepo) GetCommentByIds(ctx context.Context, ids []int64) ([]domain.Comment, error) {
	vals, err := c.dao.FindOneByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	comments := make([]domain.Comment, 0, len(vals))
	for _, v := range vals {
		comment := c.toDomain(v)
		comments = append(comments, comment)
	}
	return comments, nil
}

func (c *CachedCommentRepo) toDomain(daoComment dao.Comment) domain.Comment {
	val := domain.Comment{
		Id: daoComment.ID,
		// 数据库里面不可能存整个 User,存个ID,自己去查吧
		Commentator: domain.User{
			ID: daoComment.Uid,
		},
		Biz:     daoComment.Biz,
		BizID:   daoComment.BizID,
		Content: daoComment.Content,
		CTime:   time.UnixMilli(daoComment.Ctime),
		UTime:   time.UnixMilli(daoComment.Utime),
	}
	if daoComment.PID.Valid {
		val.ParentComment = &domain.Comment{
			Id: daoComment.PID.Int64,
		}
	}
	if daoComment.RootID.Valid {
		val.RootComment = &domain.Comment{
			Id: daoComment.RootID.Int64,
		}
	}
	return val
}

func (c *CachedCommentRepo) toEntity(domainComment domain.Comment) dao.Comment {
	daoComment := dao.Comment{
		ID:      domainComment.Id,
		Uid:     domainComment.Commentator.ID,
		Biz:     domainComment.Biz,
		BizID:   domainComment.BizID,
		Content: domainComment.Content,
	}
	if domainComment.RootComment != nil {
		daoComment.RootID = sql.NullInt64{
			Valid: true,
			Int64: domainComment.RootComment.Id,
		}
	}
	if domainComment.ParentComment != nil {
		daoComment.PID = sql.NullInt64{
			Valid: true,
			Int64: domainComment.ParentComment.Id,
		}
	}
	daoComment.Ctime = time.Now().UnixMilli()
	daoComment.Utime = time.Now().UnixMilli()
	return daoComment
}

func NewCommentRepo(commentDAO dao.CommentDAO, l logger.LoggerV1) CommentRepository {
	return &CachedCommentRepo{
		dao: commentDAO,
		l:   l,
	}
}
