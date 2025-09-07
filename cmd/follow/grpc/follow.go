package grpc

import (
	"context"
	followv1 "github.com/XD/ScholarNet/cmd/api/proto/gen/follow/v1"
	"github.com/XD/ScholarNet/cmd/follow/domain"
	"github.com/XD/ScholarNet/cmd/follow/service"
	"google.golang.org/grpc"
)

type FollowServiceServer struct {
	followv1.UnimplementedFollowServiceServer
	svc service.FollowRelationService
}

func NewFollowServiceServer(svc service.FollowRelationService) *FollowServiceServer {
	return &FollowServiceServer{
		svc: svc,
	}
}

func (f *FollowServiceServer) Register(server *grpc.Server) {
	followv1.RegisterFollowServiceServer(server, f)
}

func (f *FollowServiceServer) Follow(ctx context.Context, request *followv1.FollowRequest) (*followv1.FollowResponse, error) {
	err := f.svc.Follow(ctx, request.GetFollower(), request.GetFollowee())
	return &followv1.FollowResponse{}, err
}

func (f *FollowServiceServer) CancelFollow(ctx context.Context, request *followv1.CancelFollowRequest) (*followv1.CancelFollowResponse, error) {
	err := f.svc.CancelFollow(ctx, request.GetFollower(), request.GetFollowee())
	return &followv1.CancelFollowResponse{}, err
}

func (f *FollowServiceServer) GetFollowee(ctx context.Context, request *followv1.GetFolloweeRequest) (*followv1.GetFolloweeResponse, error) {
	relationList, err := f.svc.GetFollowee(ctx, request.GetFollower(), request.GetOffset(), request.GetLimit())
	if err != nil {
		return nil, err
	}
	res := make([]*followv1.FollowRelation, 0, len(relationList))
	for _, relation := range relationList {
		res = append(res, f.convertToView(relation))
	}
	return &followv1.GetFolloweeResponse{
		FollowRelations: res,
	}, nil
}

func (f *FollowServiceServer) FollowInfo(ctx context.Context, request *followv1.FollowInfoRequest) (*followv1.FollowInfoResponse, error) {
	relation, err := f.svc.FollowInfo(ctx, request.GetFollower(), request.GetFollowee())
	if err != nil {
		return nil, err
	}
	return &followv1.FollowInfoResponse{
		FollowRelation: f.convertToView(relation),
	}, nil
}

func (f *FollowServiceServer) GetFollowStatics(ctx context.Context, request *followv1.GetFollowStaticsRequest) (*followv1.GetFollowStaticsResponse, error) {
	static, err := f.svc.GetFollowStatus(ctx, request.GetUid())
	if err != nil {
		return nil, err
	}
	return &followv1.GetFollowStaticsResponse{
		Followers: static.Followers,
		Followees: static.Followees,
	}, err
}

func (f *FollowServiceServer) convertToView(relation domain.FollowRelation) *followv1.FollowRelation {
	return &followv1.FollowRelation{
		Follower: relation.Follower,
		Followee: relation.Followee,
	}
}
