package grpc

import (
	"context"
	articlev1 "github.com/XD/ScholarNet/cmd/api/proto/gen/article/v1"
	"github.com/XD/ScholarNet/cmd/article/domain"
	"github.com/XD/ScholarNet/cmd/article/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ArticleServiceServer struct {
	articlev1.UnimplementedArticleServiceServer
	service service.ArticleService
}

func NewArticleServiceServer(svc service.ArticleService) *ArticleServiceServer {
	return &ArticleServiceServer{
		service: svc,
	}
}
func (a *ArticleServiceServer) Register(server grpc.ServiceRegistrar) {
	articlev1.RegisterArticleServiceServer(server, a)
}

func (a *ArticleServiceServer) Save(ctx context.Context, request *articlev1.SaveRequest) (*articlev1.SaveResponse, error) {
	art, err := convertToDomain(request.Article)
	if err != nil {
		return nil, err
	}
	id, err := a.service.Save(ctx, art)
	return &articlev1.SaveResponse{Id: id}, err
}

func (a *ArticleServiceServer) Publish(ctx context.Context, request *articlev1.PublishRequest) (*articlev1.PublishResponse, error) {
	art, err := convertToDomain(request.Article)
	if err != nil {
		return nil, err
	}
	id, err := a.service.Publish(ctx, art)
	return &articlev1.PublishResponse{Id: id}, err
}

func (a *ArticleServiceServer) Withdraw(ctx context.Context, request *articlev1.WithdrawRequest) (*articlev1.WithdrawResponse, error) {
	err := a.service.Withdraw(ctx, request.GetUid(), request.GetId())
	return &articlev1.WithdrawResponse{}, err
}

func (a *ArticleServiceServer) List(ctx context.Context, request *articlev1.ListRequest) (*articlev1.ListResponse, error) {
	articleList, err := a.service.List(ctx, request.GetAuthor(), int(request.GetOffset()), int(request.GetLimit()))
	if err != nil {
		return nil, err
	}
	list := make([]*articlev1.Article, 0, len(articleList))
	for _, article := range articleList {
		newArticle, err := convertToV(article)
		if err != nil {
			return nil, err
		}
		list = append(list, newArticle)
	}
	return &articlev1.ListResponse{
		Articles: list,
	}, nil
}

func (a *ArticleServiceServer) GetById(ctx context.Context, request *articlev1.GetByIdRequest) (*articlev1.GetByIdResponse, error) {
	art, err := a.service.GetById(ctx, request.GetId())
	newArticle, err := convertToV(art)
	if err != nil {
		return nil, err
	}
	return &articlev1.GetByIdResponse{
		Article: newArticle,
	}, nil
}

func (a *ArticleServiceServer) GetPublishedById(ctx context.Context, request *articlev1.GetPublishedByIdRequest) (*articlev1.GetPublishedByIdResponse, error) {
	art, err := a.service.GetPublishedById(ctx, request.GetId(), request.GetUid())
	newArticle, err := convertToV(art)
	if err != nil {
		return nil, err
	}
	return &articlev1.GetPublishedByIdResponse{
		Article: newArticle,
	}, nil
}

func (a *ArticleServiceServer) ListPub(ctx context.Context, request *articlev1.ListPubRequest) (*articlev1.ListPubResponse, error) {
	artList, err := a.service.ListPub(ctx, request.GetStartTime().AsTime(), int(request.GetOffset()), int(request.GetLimit()))
	if err != nil {
		return nil, err
	}
	list := make([]*articlev1.Article, 0, len(artList))
	for _, art := range artList {
		newArticle, err := convertToV(art)
		if err != nil {
			return nil, err
		}
		list = append(list, newArticle)
	}
	return &articlev1.ListPubResponse{
		Articles: list,
	}, nil
}

func convertToV(domainArticle domain.Article) (*articlev1.Article, error) {
	newArticle := articlev1.Article{}
	newArticle.Id = domainArticle.Id
	newArticle.Title = domainArticle.Title
	newArticle.Status = int32(domainArticle.Status)
	newArticle.Content = domainArticle.Content
	newArticle.Author = &articlev1.Author{
		Id:   domainArticle.Author.Id,
		Name: domainArticle.Author.Name,
	}
	newArticle.Ctime = timestamppb.New(domainArticle.Ctime)
	newArticle.Utime = timestamppb.New(domainArticle.Utime)
	return &newArticle, nil
}

func convertToDomain(vArticle *articlev1.Article) (domain.Article, error) {
	art := domain.Article{}
	if vArticle != nil {
		art.Id = vArticle.GetId()
		art.Content = vArticle.GetContent()
		art.Author = domain.Author{
			Id:   vArticle.Author.GetId(),
			Name: vArticle.Author.GetName(),
		}
		art.Status = domain.ArticleStatus(vArticle.GetStatus())
		art.Title = vArticle.GetTitle()
		art.Ctime = vArticle.GetCtime().AsTime()
		art.Utime = vArticle.GetUtime().AsTime()
	}
	return art, nil
}
