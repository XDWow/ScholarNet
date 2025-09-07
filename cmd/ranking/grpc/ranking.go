package grpc

import (
	"context"
	rankingv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/ranking/v1"
	"github.com/XD/ScholarNet/cmd/ranking/domain"
	"github.com/XD/ScholarNet/cmd/ranking/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RankingServiceServer struct {
	svc service.RankingService
	rankingv1.UnimplementedRankingServiceServer
}

func NewRankingServiceServer(svc service.RankingService) *RankingServiceServer {
	return &RankingServiceServer{
		svc: svc,
	}
}

func (r *RankingServiceServer) Register(server grpc.ServiceRegistrar) {
	rankingv1.RegisterRankingServiceServer(server, r)
}

func (r *RankingServiceServer) RankTopN(ctx context.Context, request *rankingv1.RankTopNRequest) (*rankingv1.RankTopNResponse, error) {
	err := r.svc.RankTopN(ctx)
	return &rankingv1.RankTopNResponse{}, err
}

func (r *RankingServiceServer) TopN(ctx context.Context, request *rankingv1.TopNRequest) (*rankingv1.TopNResponse, error) {
	domainArticles, err := r.svc.TopN(ctx)
	if err != nil {
		return &rankingv1.TopNResponse{}, err
	}
	res := make([]*rankingv1.Article, 0, len(domainArticles))
	for _, art := range domainArticles {
		res = append(res, convertToV(art))
	}
	return &rankingv1.TopNResponse{
		Articles: res,
	}, nil
}

func convertToV(domainArticle domain.Article) *rankingv1.Article {
	return &rankingv1.Article{
		Id:      domainArticle.Id,
		Title:   domainArticle.Title,
		Status:  int32(domainArticle.Status),
		Content: domainArticle.Content,
		Author: &rankingv1.Author{
			Id:   domainArticle.Author.Id,
			Name: domainArticle.Author.Name,
		},
		Ctime: timestamppb.New(domainArticle.Ctime),
		Utime: timestamppb.New(domainArticle.Utime),
	}
}
