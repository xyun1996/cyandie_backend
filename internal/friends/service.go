package friends

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cyandie/backend/internal/core/errors"
	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	pendingStatus  = "pending"
	acceptedStatus = "accepted"
)

type FriendsService struct {
	queries db.Querier
	rdb     *redis.Client
}

func NewFriendsService(queries db.Querier, rdb *redis.Client) *FriendsService {
	return &FriendsService{queries: queries, rdb: rdb}
}

func (s *FriendsService) SendRequest(ctx context.Context, fromUserID, toUserID string) (*db.Friendship, error) {
	from, err := uuid.Parse(fromUserID)
	if err != nil {
		return nil, errors.New(errors.ErrBadRequest, "invalid from user ID")
	}
	to, err := uuid.Parse(toUserID)
	if err != nil {
		return nil, errors.New(errors.ErrBadRequest, "invalid to user ID")
	}
	if from == to {
		return nil, errors.New(errors.ErrBadRequest, "cannot send friend request to yourself")
	}

	existing, err := s.queries.GetFriendship(ctx, from)
	if err == nil {
		_ = existing
		return nil, errors.New(errors.ErrConflict, "friendship already exists")
	}

	friendship, err := s.queries.CreateFriendship(ctx, db.CreateFriendshipParams{
		UserID:   from,
		FriendID: to,
		Status:   pendingStatus,
	})
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to create friend request")
	}

	return &friendship, nil
}

func (s *FriendsService) AcceptRequest(ctx context.Context, friendshipID string) (*db.Friendship, error) {
	id, err := uuid.Parse(friendshipID)
	if err != nil {
		return nil, errors.New(errors.ErrBadRequest, "invalid friendship ID")
	}

	friendship, err := s.queries.UpdateFriendshipStatus(ctx, db.UpdateFriendshipStatusParams{
		ID:     id,
		Status: acceptedStatus,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ErrNotFound, "friend request not found")
		}
		return nil, errors.New(errors.ErrInternal, "failed to accept friend request")
	}

	return &friendship, nil
}

func (s *FriendsService) RejectRequest(ctx context.Context, friendshipID string) error {
	id, err := uuid.Parse(friendshipID)
	if err != nil {
		return errors.New(errors.ErrBadRequest, "invalid friendship ID")
	}

	_, err = s.queries.DeleteFriendship(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New(errors.ErrNotFound, "friend request not found")
		}
		return errors.New(errors.ErrInternal, "failed to reject friend request")
	}
	return nil
}

func (s *FriendsService) ListFriends(ctx context.Context, userID string) ([]db.Friendship, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New(errors.ErrBadRequest, "invalid user ID")
	}

	friendships, err := s.queries.ListFriends(ctx, id)
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to list friends")
	}
	return friendships, nil
}

func (s *FriendsService) ListPendingRequests(ctx context.Context, userID string) ([]db.Friendship, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New(errors.ErrBadRequest, "invalid user ID")
	}

	requests, err := s.queries.ListPendingRequests(ctx, id)
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to list pending requests")
	}
	return requests, nil
}

func (s *FriendsService) SetOnline(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return errors.New(errors.ErrBadRequest, "invalid user ID")
	}

	key := fmt.Sprintf("friends:online:%s", id)
	return s.rdb.Set(ctx, key, time.Now().Unix(), 5*time.Minute).Err()
}

func (s *FriendsService) SetOffline(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return errors.New(errors.ErrBadRequest, "invalid user ID")
	}

	key := fmt.Sprintf("friends:online:%s", id)
	return s.rdb.Del(ctx, key).Err()
}

func (s *FriendsService) GetOnlineFriends(ctx context.Context, userID string) ([]string, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New(errors.ErrBadRequest, "invalid user ID")
	}

	friendships, err := s.queries.ListFriends(ctx, id)
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to list friends")
	}

	var online []string
	for _, f := range friendships {
		friendID := f.FriendID.String()
		if f.FriendID == id {
			friendID = f.UserID.String()
		}
		key := fmt.Sprintf("friends:online:%s", friendID)
		if s.rdb.Exists(ctx, key).Val() == 1 {
			online = append(online, friendID)
		}
	}
	return online, nil
}
