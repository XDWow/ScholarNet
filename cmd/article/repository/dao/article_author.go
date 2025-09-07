package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

// ArticleAuthorDAO 用于演示代表
type ArticleAuthorDAO interface {
	Create(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
}

// GORMArticleAuthorDAO 复制了 GORMArticleDAO 的代码
type GORMArticleAuthorDAO struct {
	db *gorm.DB
}

func NewGORMArticleAuthorDAO(db *gorm.DB) ArticleAuthorDAO {
	return &GORMArticleAuthorDAO{
		db: db,
	}
}

func (dao *GORMArticleAuthorDAO) Create(ctx context.Context,
	art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

// UpdateById 只更新标题和
func (dao *GORMArticleAuthorDAO) UpdateById(ctx context.Context,
	art Article) error {
	now := time.Now().UnixMilli()
	res := dao.db.Model(&Article{}).WithContext(ctx).
		Where("id=? AND author_id = ? ", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
		})
	err := res.Error
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}
