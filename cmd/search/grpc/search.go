package grpc

import (
	"context"
	searchv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/search/v1"
	"github.com/XD/ScholarNet/cmd/search/domain"
	"github.com/XD/ScholarNet/cmd/search/service"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
)

type SearchServiceServer struct {
	searchv1.UnimplementedSearchServiceServer
	svc service.SearchService
}

func NewSearchService(svc service.SearchService) *SearchServiceServer {
	return &SearchServiceServer{svc: svc}
}

func (s *SearchServiceServer) Register(server grpc.ServiceRegistrar) {
	searchv1.RegisterSearchServiceServer(server, s)
}

func (s *SearchServiceServer) Search(ctx context.Context, request *searchv1.SearchRequest) (*searchv1.SearchResponse, error) {
	res, err := s.svc.Search(ctx, request.GetUid(), request.GetExpression())
	if err != nil {
		return nil, err
	}
	return &searchv1.SearchResponse{
		User: slice.Map(res.Users, func(idx int, src domain.User) *searchv1.User {
			return &searchv1.User{
				Id:       src.Id,
				Nickname: src.Nickname,
				Email:    src.Email,
				Phone:    src.Phone,
			}
		}),
		Article: slice.Map(res.Articles, func(idx int, src domain.Article) *searchv1.Article {
			return &searchv1.Article{
				Id:      src.Id,
				Title:   src.Title,
				Status:  src.Status,
				Content: src.Content,
				Tags:    src.Tags,
			}
		}),
	}, nil
}
