package grpc

import (
	"context"
	tagv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/tag/v1"
	"github.com/XD/ScholarNet/cmd/tag/domain"
	"github.com/XD/ScholarNet/cmd/tag/service"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
)

var _ tagv1.TagServiceServer = (*TagServiceServer)(nil)

type TagServiceServer struct {
	tagv1.UnimplementedTagServiceServer
	service service.TagService
}

func (t *TagServiceServer) Register(server grpc.ServiceRegistrar) {
	tagv1.RegisterTagServiceServer(server, t)
}

func (t *TagServiceServer) CreateTag(ctx context.Context, request *tagv1.CreateTagRequest) (*tagv1.CreateTagResponse, error) {
	id, err := t.service.CreateTag(ctx, request.GetUid(), request.GetName())
	return &tagv1.CreateTagResponse{
		Tag: &tagv1.Tag{
			Id:   id,
			Uid:  request.Uid,
			Name: request.Name,
		},
	}, err
}

func (t *TagServiceServer) AttachTags(ctx context.Context, request *tagv1.AttachTagsRequest) (*tagv1.AttachTagsResponse, error) {
	err := t.service.AttachTags(ctx, request.Uid, request.Biz, request.BizId, request.Tids)
	return &tagv1.AttachTagsResponse{}, err
}

func (t *TagServiceServer) GetTags(ctx context.Context, request *tagv1.GetTagsRequest) (*tagv1.GetTagsResponse, error) {
	tags, err := t.service.GetTags(ctx, request.GetUid())
	if err != nil {
		return nil, err
	}
	return &tagv1.GetTagsResponse{
		Tag: slice.Map(tags, func(idx int, src domain.Tag) *tagv1.Tag {
			return t.toDTO(src)
		}),
	}, nil
}

func (t *TagServiceServer) GetBizTags(ctx context.Context, req *tagv1.GetBizTagsRequest) (*tagv1.GetBizTagsResponse, error) {
	res, err := t.service.GetBizTags(ctx, req.Uid, req.Biz, req.BizId)
	if err != nil {
		return nil, err
	}
	return &tagv1.GetBizTagsResponse{
		Tags: slice.Map(res, func(idx int, src domain.Tag) *tagv1.Tag {
			return t.toDTO(src)
		}),
	}, nil
}

func (t *TagServiceServer) toDTO(tag domain.Tag) *tagv1.Tag {
	return &tagv1.Tag{
		Id:   tag.Id,
		Uid:  tag.Uid,
		Name: tag.Name,
	}
}

func NewTagServiceServer(svc service.TagService) *TagServiceServer {
	return &TagServiceServer{
		service: svc,
	}
}
