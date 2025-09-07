package dao

import (
	"context"
	"github.com/olivere/elastic/v7"
)

type anyElasticDAO struct {
	client *elastic.Client
}

func NewAnyElasticDAO(client *elastic.Client) AnyDAO {
	return &anyElasticDAO{client}
}

func (dao *anyElasticDAO) Insert(ctx context.Context, index, docID, data string) error {
	// 直接整个 data 从 Kafka/grpc 里面一路透传到这里
	_, err := dao.client.Index().Index(index).Id(docID).BodyJson(data).Do(ctx)
	return err
}
