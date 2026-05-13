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

// PresenceNotifier pushes real-time social events to users.
// Defined locally to avoid circular import with chat package.
type PresenceNotifier interface {
	NotifyOnline(ctx context.Context, userID, username string, friendIDs []string)
	NotifyOffline(ctx context.Context, userID string, friendIDs []string)
	NotifyFriendRemoved(ctx context.Context, targetUserID, removedByID string)
	NotifyBlocked(ctx context.Context, blockedUserID, blockerID string)
}

type FriendsService struct {
	queries  db.Querier
	rdb      *redis.Client
	notifier PresenceNotifier
}

func NewFriendsService(queries db.Querier, rdb *redis.Client, notifier PresenceNotifier) *FriendsService {
	return &FriendsService{queries: queries, rdb: rdb, notifier: notifier}
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

	blocked, err := s.IsBlocked(ctx, toUserID, fromUserID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, errors.New(errors.ErrForbidden, "you are blocked by this user")
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

func (s *FriendsService) SetOnline(ctx context.Context, userID, username string) error {
	if s.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("friends:online:%s", userID)
	// SETNX returns true if the key was set (first online)
	set, err := s.rdb.SetNX(ctx, key, username, 5*time.Minute).Result()
	if err != nil {
		return errors.New(errors.ErrInternal, "failed to set online")
	}
	if set && s.notifier != nil {
		// Just came online — notify friends
		friendships, _ := s.queries.ListFriends(ctx, uuid.MustParse(userID))
		friendIDs := make([]string, 0, len(friendships))
		for _, f := range friendships {
			friendIDs = append(friendIDs, f.FriendID.String())
		}
		s.notifier.NotifyOnline(ctx, userID, username, friendIDs)
	}
	// Refresh TTL
	s.rdb.Expire(ctx, key, 5*time.Minute)
	return nil
}

func (s *FriendsService) SetOffline(ctx context.Context, userID string) error {
	if s.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("friends:online:%s", userID)
	// Check if currently online before removing
	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return errors.New(errors.ErrInternal, "failed to check online")
	}
	s.rdb.Del(ctx, key)
	if exists > 0 && s.notifier != nil {
		friendships, _ := s.queries.ListFriends(ctx, uuid.MustParse(userID))
		friendIDs := make([]string, 0, len(friendships))
		for _, f := range friendships {
			friendIDs = append(friendIDs, f.FriendID.String())
		}
		s.notifier.NotifyOffline(ctx, userID, friendIDs)
	}
	return nil
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

func (s *FriendsService) Block(ctx context.Context, blockerID, blockedID, reason string) error {
	bid, err := uuid.Parse(blockerID)
	if err != nil {
		return errors.New(errors.ErrBadRequest, "invalid blocker id")
	}
	blid, err := uuid.Parse(blockedID)
	if err != nil {
		return errors.New(errors.ErrBadRequest, "invalid blocked id")
	}
	if blockerID == blockedID {
		return errors.New(errors.ErrBadRequest, "cannot block yourself")
	}

	_, err = s.queries.CreateBlockRelation(ctx, db.CreateBlockRelationParams{
		BlockerID: bid,
		BlockedID: blid,
		Reason:    sql.NullString{String: reason, Valid: reason != ""},
	})
	if err != nil {
		return errors.New(errors.ErrInternal, "failed to create block relation")
	}

	// Delete friendship if exists
	_, _ = s.queries.DeleteFriendshipByUsers(ctx, db.DeleteFriendshipByUsersParams{
		UserID:   bid,
		FriendID: blid,
	})

	// Reject pending friend requests involving both users
	pending, _ := s.queries.ListPendingRequests(ctx, bid)
	for _, f := range pending {
		if f.Status == pendingStatus && (f.FriendID == blid || f.UserID == blid) {
			s.queries.UpdateFriendshipStatus(ctx, db.UpdateFriendshipStatusParams{
				ID:     f.ID,
				Status: "rejected",
			})
		}
	}

	// Update Redis cache
	if s.rdb != nil {
		s.rdb.SAdd(ctx, fmt.Sprintf("friends:blocked:%s", blockerID), blockedID)
	}

	// Notify blocked user
	if s.notifier != nil {
		s.notifier.NotifyBlocked(ctx, blockedID, blockerID)
	}

	return nil
}

func (s *FriendsService) Unblock(ctx context.Context, blockerID, blockedID string) error {
	bid, err := uuid.Parse(blockerID)
	if err != nil {
		return errors.New(errors.ErrBadRequest, "invalid blocker id")
	}
	blid, err := uuid.Parse(blockedID)
	if err != nil {
		return errors.New(errors.ErrBadRequest, "invalid blocked id")
	}

	_, err = s.queries.DeleteBlockRelation(ctx, db.DeleteBlockRelationParams{
		BlockerID: bid,
		BlockedID: blid,
	})
	if err != nil {
		return errors.New(errors.ErrInternal, "failed to delete block relation")
	}

	if s.rdb != nil {
		s.rdb.SRem(ctx, fmt.Sprintf("friends:blocked:%s", blockerID), blockedID)
	}

	return nil
}

func (s *FriendsService) IsBlocked(ctx context.Context, targetUserID, byUserID string) (bool, error) {
	// Check Redis first
	if s.rdb != nil {
		isMember, err := s.rdb.SIsMember(ctx, fmt.Sprintf("friends:blocked:%s", targetUserID), byUserID).Result()
		if err == nil {
			return isMember, nil
		}
	}

	// Fallback to DB
	tid, err := uuid.Parse(targetUserID)
	if err != nil {
		return false, errors.New(errors.ErrBadRequest, "invalid target user id")
	}
	bid, err := uuid.Parse(byUserID)
	if err != nil {
		return false, errors.New(errors.ErrBadRequest, "invalid by user id")
	}

	_, err = s.queries.IsBlockedBy(ctx, db.IsBlockedByParams{
		BlockerID: tid,
		BlockedID: bid,
	})
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, errors.New(errors.ErrInternal, "failed to check block relation")
	}
	return true, nil
}

func (s *FriendsService) ListBlockedUsers(ctx context.Context, userID string) ([]db.BlockRelation, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New(errors.ErrBadRequest, "invalid user id")
	}
	return s.queries.ListBlockedUsers(ctx, uid)
}

// RemoveFriend deletes the friendship between two users and notifies the other party.
func (s *FriendsService) RemoveFriend(ctx context.Context, userID, friendID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New(errors.ErrBadRequest, "invalid user id")
	}
	fid, err := uuid.Parse(friendID)
	if err != nil {
		return errors.New(errors.ErrBadRequest, "invalid friend id")
	}

	_, err = s.queries.DeleteFriendshipByUsers(ctx, db.DeleteFriendshipByUsersParams{
		UserID:   uid,
		FriendID: fid,
	})
	if err != nil {
		return errors.New(errors.ErrInternal, "failed to delete friendship")
	}

	if s.notifier != nil {
		s.notifier.NotifyFriendRemoved(ctx, friendID, userID)
	}

	return nil
}

// ListRecentContacts returns recent contacts sorted by last interaction time.
func (s *FriendsService) ListRecentContacts(ctx context.Context, userID string, limit int) ([]string, error) {
	if s.rdb == nil {
		return nil, errors.New(errors.ErrInternal, "redis not available")
	}

	key := fmt.Sprintf("friends:recent:%s", userID)
	now := float64(time.Now().Unix())

	// Remove entries older than 30 days
	cutoff := now - 30*24*3600
	s.rdb.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%f", cutoff))

	results, err := s.rdb.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Max:   fmt.Sprintf("%f", now),
		Min:   "0",
		Count: int64(limit),
	}).Result()
	if err != nil {
		return nil, errors.New(errors.ErrInternal, "failed to get recent contacts")
	}

	return results, nil
}

// PlatformFriendInfo represents a friend from an external platform.
type PlatformFriendInfo struct {
	PlatformUserID string
	Username       string
}

// ImportPlatformFriends is a stub for future platform friend import.
func (s *FriendsService) ImportPlatformFriends(_ context.Context, _, _ string, _ []PlatformFriendInfo) error {
	return errors.New(errors.ErrNotImplemented, "platform friend import is not yet available")
}
