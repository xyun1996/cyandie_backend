package admin

import (
	"context"
	"database/sql"

	"github.com/cyandie/backend/internal/core/errors"
	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	"golang.org/x/crypto/bcrypt"
)

type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminService struct {
	queries db.Querier
}

func NewAdminService(queries db.Querier) *AdminService {
	return &AdminService{queries: queries}
}

func (s *AdminService) Login(ctx context.Context, req AdminLoginRequest) (*db.AdminUser, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.New(errors.ErrInvalidCredentials, "username and password are required")
	}

	admin, err := s.queries.GetAdminByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New(errors.ErrInvalidCredentials, "invalid credentials")
	}

	if admin.Status != "active" {
		return nil, errors.New(errors.ErrForbidden, "admin account is not active")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New(errors.ErrInvalidCredentials, "invalid credentials")
	}

	return &admin, nil
}

func (s *AdminService) ListUsers(ctx context.Context, limit, offset int32) ([]db.User, error) {
	users, err := s.queries.SearchUsers(ctx, db.SearchUsersParams{
		Column1: "%",
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to list users")
	}
	return users, nil
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, userID string, status string) (*db.User, error) {
	if status != "active" && status != "banned" && status != "deleted" {
		return nil, errors.New(errors.ErrBadRequest, "invalid status: must be active, banned, or deleted")
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New(errors.ErrBadRequest, "invalid user ID")
	}

	user, err := s.queries.UpdateUserStatus(ctx, db.UpdateUserStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ErrNotFound, "user not found")
		}
		return nil, errors.New(errors.ErrInternal, "failed to update user status")
	}
	return &user, nil
}

func (s *AdminService) ListAuditLogs(ctx context.Context, limit, offset int32) ([]db.AuditLog, error) {
	logs, err := s.queries.ListAuditLogs(ctx, db.ListAuditLogsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to list audit logs")
	}
	return logs, nil
}

func (s *AdminService) CreateAuditLog(ctx context.Context, operatorID, action, targetType, targetID, reason, ip string) (*db.AuditLog, error) {
	opID, _ := uuid.Parse(operatorID)

	log, err := s.queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		OperatorID:  uuid.NullUUID{UUID: opID, Valid: opID != uuid.Nil},
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		BeforeValue: pqtype.NullRawMessage{},
		AfterValue:  pqtype.NullRawMessage{},
		Reason:      sql.NullString{String: reason, Valid: reason != ""},
		Ip:          sql.NullString{String: ip, Valid: ip != ""},
	})
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to create audit log")
	}
	return &log, nil
}
