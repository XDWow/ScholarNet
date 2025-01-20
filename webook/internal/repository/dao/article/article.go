package article

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	// 制作库
	Insert(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	// 线上库
	Upsert(ctx context.Context, article Article) error
	Sync(ctx context.Context, article Article) (int64, error)
	Transaction(ctx context.Context,
		bizFunc func(txDAO ArticleDAO) error) error
	SyncStatus(ctx context.Context, id int64, author int64, status uint8) error
}

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.Transaction(func(tx *gorm.DB) error {
		// 这个 func 是一个事务
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", id, author).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			})
		// 数据库出问题
		if res.Error != nil {
			return res.Error
		}
		// 数据库没问题，但是没找到，即为 id 跟 author_id 没对上
		if res.RowsAffected == 0 {
			return fmt.Errorf("可能有人在搞你，去操作别人的文章， uid: %d, aid: %d", author, id)
		}

		// 修改完制作库，再修改线上库，不需要再验证 author_id 了，前面已经验证过了
		return tx.Model(&Article{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status": status,
			"utime":  now,
		}).Error
	})
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now
	// GORM 会将结构体的名字（例如 article）转化为对应的表名
	// 这里就是操作 article 这张表
	err := dao.db.WithContext(ctx).Create(&article).Error
	return article.Id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	article.Utime = now
	// 依赖 gorm 忽略零值的特性，会用主键进行更新，不过可读性很差，不建议用
	// 这样实习僧也能看懂
	res := dao.db.WithContext(ctx).Model(&article).
		Where("id=? AND author_id = ?", article.Id, article.AuthorId).
		Updates(map[string]any{
			"title":   article.Title,
			"content": article.Content,
			"utime":   article.Utime,
		})
	// 要不要检查一下真的更新没有？
	// res.RowsAffected // 更新行数
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// 肯定出问题了，打日志
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d, no rows affected", article.Id, article.AuthorId)
	}
	return nil
}

// Upsert INSERT OR UPDATE
func (dao *GORMArticleDAO) Upsert(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	// 插入,	OnConflict 的意思是数据冲突了
	err := dao.db.Clauses(clause.OnConflict{
		// SQL 2003 标准
		// INSERT AAAA ON CONFLICT(BBB) DO NOTHING
		// INSERT AAAA ON CONFLICT(BBB) DO UPDATES CCC WHERE DDD

		// 哪些列冲突
		//Columns: []clause.Column{clause.Column{Name: "id"}},
		// 意思是数据冲突，啥也不干
		// DoNothing:
		// 数据冲突了，并且符合 WHERE 条件的就会执行 DO UPDATES
		// Where:

		// MySQL 只需要关心这里，它不是 2003 标准
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   article.Title,
			"content": article.Content,
			"utime":   now,
		}),
	}).Create(&article).Error
	// MySQL 最终的语句是 INSERT xxx ON DUPLICATE KEY UPDATE xxx
	return err
}

func (dao *GORMArticleDAO) Transaction(ctx context.Context,
	bizFunc func(txDAO ArticleDAO) error) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		txDAO := NewGORMArticleDAO(tx)
		return bizFunc(txDAO)
	})
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 先操作制作库（此时应该是表），后操作线上库（此时应该是表）
	var (
		id = art.Id
	)
	// tx => Transaction, trx, txn
	// 在事务内部，这里采用了闭包形态
	// GORM 帮助我们管理了事务的生命周期
	// Begin，Rollback 和 Commit 都不需要我们操心
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error
		txDAO := NewGORMArticleDAO(tx)
		if id > 0 {
			err = txDAO.UpdateById(ctx, art)
		} else {
			id, err = txDAO.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		// 同步到线上库
		return txDAO.Upsert(ctx, art)
	})
	return id, err
}

type Article struct {
	Id    int64  `gorm:"primary_key;auto_increment"`
	Title string `gorm:"type=varchar(1024)"`
	// BLOB 类型在数据库中通常用于存储大量的二进制数据,比如文件、图像、音频文件等
	// 这类数据不适合存储为常规的文本类型（如 VARCHAR 或 TEXT），因为它们可能包含无法通过文本处理的字符。
	Content string `gorm:"type=BLOB"`
	// 如何设计索引
	// 在帖子这里，什么样查询场景？
	// 对于创作者来说，是不是看草稿箱，看到所有自己的文章？
	// SELECT * FROM articles WHERE author_id = 123 ORDER BY `ctime` DESC;
	// 产品经理告诉你，要按照创建时间的倒序排序
	// 单独查询某一篇 SELECT * FROM articles WHERE id = 1
	// 在查询接口，我们深入讨论这个问题
	// - 在 author_id 和 ctime 上创建联合索引,这样查找并按ctime排序返回的效率更高
	// - 在 author_id 上创建索引

	// 学学 Explain 命令

	// 在 author_id 上创建索引
	AuthorId int64 `gorm:"index"`

	// 有些人考虑到，经常用状态来查询
	// WHERE status = xxx AND
	// 在 status 上和别的列混在一起，创建一个联合索引
	// 要看别的列究竟是什么列。
	Status uint8
	Ctime  int64
	Utime  int64
}
