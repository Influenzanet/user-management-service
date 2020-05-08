package service

import (
	"context"
	"log"
	"time"

	"github.com/influenzanet/user-management-service/pkg/api"
	"github.com/influenzanet/user-management-service/pkg/models"
	"github.com/influenzanet/user-management-service/pkg/tokens"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *userManagementServer) GenerateTempToken(ctx context.Context, t *api.TempTokenInfo) (*api.TempToken, error) {
	if t == nil || t.Purpose == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	tempToken := models.TempToken{
		UserID:     t.UserId,
		InstanceID: t.InstanceId,
		Purpose:    t.Purpose,
		Info:       t.Info,
		Expiration: t.Expiration,
	}

	if tempToken.Expiration == 0 {
		tempToken.Expiration = tokens.GetExpirationTime(time.Hour * 24 * 10)
	}

	token, err := s.globalDBService.AddTempToken(tempToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.TempToken{
		Token: token,
	}, nil
}

func (s *userManagementServer) ValidateTempToken(ctx context.Context, t *api.TempToken) (*api.TempTokenInfo, error) {
	if t == nil || t.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	tempToken, err := s.globalDBService.GetTempToken(t.Token)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if time.Now().Unix() > tempToken.Expiration {
		err = s.globalDBService.DeleteTempToken(tempToken.Token)
		log.Println(err)
		return nil, status.Error(codes.InvalidArgument, "token expired")
	}

	return tempToken.ToAPI(), nil
}

func (s *userManagementServer) GetTempTokens(ctx context.Context, t *api.TempTokenInfo) (*api.TempTokenInfos, error) {
	if t == nil || t.UserId == "" || t.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}

	tokens, err := s.globalDBService.GetTempTokenForUser(t.InstanceId, t.UserId, t.Purpose)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return tokens.ToAPI(), nil
}

func (s *userManagementServer) DeleteTempToken(ctx context.Context, t *api.TempToken) (*api.Status, error) {
	if t == nil || t.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	if err := s.globalDBService.DeleteTempToken(t.Token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.ServiceStatus{
		Status: api.ServiceStatus_NORMAL,
		Msg:    "deleted",
	}, nil
}

func (s *userManagementServer) PurgeUserTempTokens(ctx context.Context, t *api.TempTokenInfo) (*api.Status, error) {
	if t == nil || t.UserId == "" || t.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing argument")
	}
	if err := deleteAllTempTokenForUserDB(t.InstanceId, t.UserId, t.Purpose); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &api.Status{
		Status: api.Status_NORMAL,
		Msg:    "deleted",
	}, nil
}
