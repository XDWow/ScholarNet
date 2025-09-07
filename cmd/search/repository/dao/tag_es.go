package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
)

type tagElasticDAO struct {
	client *elastic.Client
}

func NewTagElasticDAO(client *elastic.Client) TagDAO {
	return &tagElasticDAO{client: client}
}

func (dao *tagElasticDAO) Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("uid", uid),
		elastic.NewTermQuery("biz", biz),
		elastic.NewTermsQueryFromStrings("tags", keywords...),
	)
	resp, err := dao.client.Search(TagIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var ele BizTags
		// 这就是倒排索引
		// 根据 uid int64, biz string, keywords []string 查出 BizTags（根据属性值查数据）
		err = json.Unmarshal(hit.Source, &ele)
		if err != nil {
			return nil, err
		}
		res = append(res, ele.BizId)
	}
	return res, nil
}
