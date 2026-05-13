package chat

import (
	"context"

	chatv1 "github.com/cyandie/backend/api/proto/chat/v1"
)

// PresenceNotifier pushes real-time social events to users over TCP.
type PresenceNotifier interface {
	NotifyOnline(ctx context.Context, userID, username string, friendIDs []string)
	NotifyOffline(ctx context.Context, userID string, friendIDs []string)
	NotifyFriendRemoved(ctx context.Context, targetUserID, removedByID string)
	NotifyBlocked(ctx context.Context, blockedUserID, blockerID string)
}

// ChatPresenceNotifier implements PresenceNotifier using the TCP server.
type ChatPresenceNotifier struct {
	srv *TCPServer
}

func NewChatPresenceNotifier(srv *TCPServer) *ChatPresenceNotifier {
	return &ChatPresenceNotifier{srv: srv}
}

func (n *ChatPresenceNotifier) NotifyOnline(_ context.Context, userID, username string, friendIDs []string) {
	frame := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_PRESENCE_ONLINE,
		Payload: &chatv1.ChatEnvelope_PresenceOnline{
			PresenceOnline: &chatv1.PresenceOnline{
				UserId:   userID,
				Username: username,
			},
		},
	}
	for _, fid := range friendIDs {
		n.srv.SendToUser(fid, frame)
	}
}

func (n *ChatPresenceNotifier) NotifyOffline(_ context.Context, userID string, friendIDs []string) {
	frame := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_PRESENCE_OFFLINE,
		Payload: &chatv1.ChatEnvelope_PresenceOffline{
			PresenceOffline: &chatv1.PresenceOffline{
				UserId: userID,
			},
		},
	}
	for _, fid := range friendIDs {
		n.srv.SendToUser(fid, frame)
	}
}

func (n *ChatPresenceNotifier) NotifyFriendRemoved(_ context.Context, targetUserID, removedByID string) {
	frame := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_FRIEND_REMOVED,
		Payload: &chatv1.ChatEnvelope_FriendRemoved{
			FriendRemoved: &chatv1.FriendRemoved{
				UserId: removedByID,
			},
		},
	}
	n.srv.SendToUser(targetUserID, frame)
}

func (n *ChatPresenceNotifier) NotifyBlocked(_ context.Context, blockedUserID, blockerID string) {
	frame := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_BLOCK_NOTIFY,
		Payload: &chatv1.ChatEnvelope_BlockNotify{
			BlockNotify: &chatv1.BlockNotify{
				BlockerId: blockerID,
			},
		},
	}
	n.srv.SendToUser(blockedUserID, frame)
}
