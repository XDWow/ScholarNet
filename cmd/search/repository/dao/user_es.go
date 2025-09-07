package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

type userElasticDAO struct {
	client *elastic.Client
}

func (u *userElasticDAO) InputUser(ctx context.Context, user User) error {
	_, err := u.client.Index().Index(UserIndexName).
		Id(strconv.FormatInt(user.Id, 10)).
		BodyJson(user).Do(ctx)
	return err
}

func (u *userElasticDAO) Search(ctx context.Context, keywords []string) ([]User, error) {
	// 纯粹是因为前面我们已经预处理了输入
	queryString := strings.Join(keywords, " ")
	// 昵称命中就可以的
	resp, err := u.client.Search(UserIndexName).
		Query(elastic.NewMatchQuery("nickname", queryString)).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]User, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var user User
		err = json.Unmarshal(hit.Source, &user)
		if err != nil {
			return nil, err
		}
		res = append(res, user)
	}
	return res, nil
}

func NewUserElasticDAO(client *elastic.Client) UserDAO {
	return &userElasticDAO{client}
}
