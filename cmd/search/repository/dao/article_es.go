package dao

import (
	"context"
	"encoding/json"
	"github.com/ecodeclub/ekit/slice"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

type articleElasticDAO struct {
	client *elastic.Client
}

func (a *articleElasticDAO) InputArticle(ctx context.Context, art Article) error {
	_, err := a.client.Index().Index(ArticleIndexName).
		// 为什么要指定 ID？
		// 确保后面文章更新的时候，我们这里不是产生类似的两条数据，而是更新了数据
		Id(strconv.FormatInt(art.Id, 10)).BodyJson(art).Do(ctx)
	return err
}

func (a *articleElasticDAO) Search(ctx context.Context, ids []int64, keywords []string) ([]Article, error) {
	queryString := strings.Join(keywords, " ")
	// 文章，标题或者内容任何一个匹配上
	// 并且状态 status 必须是已发表的状态

	// status 精确查找
	statusTerm := elastic.NewTermQuery("status", 2)

	// 标签命中
	// 整个切片类型之间没有这种隐式转换 int64->any，必须手动转
	IdAnys := slice.Map(ids, func(idx int, src int64) any {
		// 单个值会隐私转换
		return src
	})

	// 内容或者标题，模糊查找（match）
	titleMatch := elastic.NewMatchQuery("title", queryString)
	contentMatch := elastic.NewMatchQuery("content", queryString)
	or := elastic.NewBoolQuery().Should(titleMatch, contentMatch)
	if len(IdAnys) > 0 {
		// 以前版本查询值为空：{ "terms": { "id": null } } ，ES会报错，现在修复了，但为了版本差异带来的坑，兼容一下
		tag := elastic.NewTermsQuery("id", IdAnys...).Boost(2.0)
		or = or.Should(tag)
	}
	and := elastic.NewBoolQuery().Must(or, statusTerm)

	//	可以抽象一个方法出来：return NewSearcher[Article](h.client, ArticleIndexName).
	//	Query(and).Do(ctx)
	resp, err := a.client.Search(ArticleIndexName).Query(and).Do(ctx)
	if err != nil {
		return nil, err
	}
	var res []Article
	for _, hit := range resp.Hits.Hits {
		var art Article
		err = json.Unmarshal(hit.Source, &art)
		if err != nil {
			return nil, err
		}
		res = append(res, art)
	}
	return res, nil
}

func NewArticleElasticDAO(client *elastic.Client) ArticleDAO {
	return &articleElasticDAO{client}
}

// 抽象方法实现

type Searcher[T any] struct {
	client  *elastic.Client
	idxName []string
	query   elastic.Query
}

func NewSearcher[T any](client *elastic.Client, idxName ...string) *Searcher[T] {
	return &Searcher[T]{
		client:  client,
		idxName: idxName,
	}
}

func (s *Searcher[T]) Query(q elastic.Query) *Searcher[T] {
	s.query = q
	return s
}

func (s *Searcher[T]) Do(ctx context.Context) ([]T, error) {
	resp, err := s.client.Search(s.idxName...).Do(ctx)
	res := make([]T, 0, resp.Hits.TotalHits.Value)
	for _, hit := range resp.Hits.Hits {
		var t T
		err = json.Unmarshal(hit.Source, &t)
		if err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}
