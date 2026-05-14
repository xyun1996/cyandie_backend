package chat

import (
	"context"
	"log/slog"

	chatv1 "github.com/cyandie/backend/api/proto/chat/v1"
	"github.com/cyandie/backend/internal/db"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

// BlockChecker checks if a user is blocked by another.
// Defined locally to avoid circular import with friends package.
type BlockChecker interface {
	IsBlocked(ctx context.Context, targetUserID, byUserID string) (bool, error)
}

// PresenceSetter updates online/offline status for presence.
// Defined locally to avoid circular import with friends package.
type PresenceSetter interface {
	SetOnline(ctx context.Context, userID, username string) error
	SetOffline(ctx context.Context, userID string) error
}

// TokenValidator validates JWT tokens and returns the user ID.
// Defined locally to avoid circular import with auth package.
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (userID string, err error)
}

type ChatService struct {
	queries         db.Querier
	server          *TCPServer
	blockChecker    BlockChecker
	presenceSetter  PresenceSetter
	tokenValidator  TokenValidator
}

func NewChatService(queries db.Querier, server *TCPServer, blockChecker BlockChecker, presenceSetter PresenceSetter, tokenValidator TokenValidator) *ChatService {
	return &ChatService{queries: queries, server: server, blockChecker: blockChecker, presenceSetter: presenceSetter, tokenValidator: tokenValidator}
}

func (s *ChatService) setBlockChecker(checker BlockChecker) {
	s.blockChecker = checker
}

func (s *ChatService) Start() error {
	s.server.OnConnect(s.handleConnect)
	s.server.OnDisconnect(s.handleDisconnect)
	s.server.OnMessage(s.handleMessage)
	return s.server.Start()
}

func (s *ChatService) Stop() error { return s.server.Close() }

func (s *ChatService) handleConnect(conn *Connection) {
	slog.Info("client connected", "conn_id", conn.ID)
}

func (s *ChatService) handleDisconnect(conn *Connection) {
	slog.Info("client disconnected", "conn_id", conn.ID, "user_id", conn.UserID)
	if conn.UserID != "" && s.presenceSetter != nil {
		s.presenceSetter.SetOffline(context.Background(), conn.UserID)
	}
}

func (s *ChatService) handleMessage(conn *Connection, frame *Frame) {
	env := &chatv1.ChatEnvelope{}
	if err := proto.Unmarshal(frame.Value, env); err != nil {
		slog.Error("unmarshal message", "error", err)
		return
	}

	switch env.Type {
	case chatv1.MessageType_AUTH:
		s.handleAuth(conn, env)
	case chatv1.MessageType_HEARTBEAT:
		conn.Send(Frame{Type: uint16(chatv1.MessageType_HEARTBEAT_ACK)})
		if conn.UserID != "" && s.presenceSetter != nil {
			s.presenceSetter.SetOnline(context.Background(), conn.UserID, conn.Username)
		}
	case chatv1.MessageType_JOIN_ROOM:
		s.handleJoinRoom(conn, env)
	case chatv1.MessageType_LEAVE_ROOM:
		s.handleLeaveRoom(conn, env)
	case chatv1.MessageType_SEND_MSG:
		s.handleSendMessage(conn, env)
	case chatv1.MessageType_INVITE_ROOM:
		s.handleInviteRoom(conn, env)
	}
}

func (s *ChatService) handleAuth(conn *Connection, env *chatv1.ChatEnvelope) {
	req := env.GetAuthRequest()
	if req == nil || req.AccessToken == "" {
		s.sendError(conn, "BAD_REQUEST", "missing access token")
		return
	}

	if s.tokenValidator != nil {
		userID, err := s.tokenValidator.ValidateToken(context.Background(), req.AccessToken)
		if err != nil {
			s.sendError(conn, "AUTH_REJECTED", "invalid token")
			return
		}
		conn.UserID = userID
	} else {
		// Fallback: accept token as user_id (for testing)
		conn.UserID = req.AccessToken
	}
	conn.Username = req.Username

	resp := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_AUTH_OK,
		Payload: &chatv1.ChatEnvelope_AuthResponse{
			AuthResponse: &chatv1.AuthResponse{UserId: conn.UserID},
		},
	}
	s.sendEnvelope(conn, resp)
}

func (s *ChatService) handleJoinRoom(conn *Connection, env *chatv1.ChatEnvelope) {
	if conn.UserID == "" {
		s.sendError(conn, "UNAUTHORIZED", "not authenticated")
		return
	}
	req := env.GetJoinRoomRequest()
	if req == nil {
		s.sendError(conn, "BAD_REQUEST", "missing room_id")
		return
	}

	uid, _ := uuid.Parse(conn.UserID)
	roomUID, _ := uuid.Parse(req.RoomId)
	_, err := s.queries.AddRoomMember(context.Background(), db.AddRoomMemberParams{
		RoomID: roomUID, UserID: uid, Role: "member",
	})
	if err != nil {
		s.sendError(conn, "INTERNAL_ERROR", "failed to join room")
		return
	}

	resp := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_JOIN_ROOM_OK,
		Payload: &chatv1.ChatEnvelope_JoinRoomResponse{
			JoinRoomResponse: &chatv1.JoinRoomResponse{RoomId: req.RoomId},
		},
	}
	s.sendEnvelope(conn, resp)
	conn.JoinRoom(req.RoomId)
}

func (s *ChatService) handleLeaveRoom(conn *Connection, env *chatv1.ChatEnvelope) {
	if conn.UserID == "" {
		s.sendError(conn, "UNAUTHORIZED", "not authenticated")
		return
	}
	req := env.GetLeaveRoomRequest()
	if req == nil {
		s.sendError(conn, "BAD_REQUEST", "missing room_id")
		return
	}

	uid, _ := uuid.Parse(conn.UserID)
	roomUID, _ := uuid.Parse(req.RoomId)
	_, err := s.queries.RemoveRoomMember(context.Background(), db.RemoveRoomMemberParams{
		RoomID: roomUID, UserID: uid,
	})
	if err != nil {
		s.sendError(conn, "INTERNAL_ERROR", "failed to leave room")
		return
	}

	conn.LeaveRoom(req.RoomId)

	resp := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_LEAVE_ROOM_OK,
		Payload: &chatv1.ChatEnvelope_LeaveRoomResponse{
			LeaveRoomResponse: &chatv1.LeaveRoomResponse{RoomId: req.RoomId},
		},
	}
	s.sendEnvelope(conn, resp)
}

func (s *ChatService) handleSendMessage(conn *Connection, env *chatv1.ChatEnvelope) {
	if conn.UserID == "" {
		s.sendError(conn, "UNAUTHORIZED", "not authenticated")
		return
	}
	req := env.GetSendMessageRequest()
	if req == nil {
		s.sendError(conn, "BAD_REQUEST", "missing message data")
		return
	}

	roomUID, err := uuid.Parse(req.RoomId)
	if err != nil {
		s.sendError(conn, "BAD_REQUEST", "invalid room id")
		return
	}

	// Verify sender is in the room
	if !conn.IsInRoom(req.RoomId) {
		s.sendError(conn, "FORBIDDEN", "you are not in this room")
		return
	}

	// Check if any room member has blocked the sender
	if s.blockChecker != nil {
		members, _ := s.queries.GetRoomMembers(context.Background(), roomUID)
		for _, m := range members {
			if m.UserID.String() == conn.UserID {
				continue
			}
			blocked, _ := s.blockChecker.IsBlocked(context.Background(), conn.UserID, m.UserID.String())
			if blocked {
				s.sendError(conn, "FORBIDDEN", "you are blocked by a room member")
				return
			}
		}
	}

	uid, _ := uuid.Parse(conn.UserID)
	msg, err := s.queries.CreateChatMessage(context.Background(), db.CreateChatMessageParams{
		RoomID:   roomUID,
		SenderID: uid,
		Content:  req.Content,
		Type:     req.Type,
	})
	if err != nil {
		s.sendError(conn, "INTERNAL_ERROR", "failed to save message")
		return
	}

	recv := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_RECV_MSG,
		Payload: &chatv1.ChatEnvelope_ReceiveMessage{
			ReceiveMessage: &chatv1.ReceiveMessage{
				RoomId:    req.RoomId,
				SenderId:  conn.UserID,
				Content:   req.Content,
				Type:      req.Type,
				Timestamp: msg.CreatedAt.Unix(),
				MessageId: msg.ID.String(),
			},
		},
	}
	s.broadcastEnvelope(req.RoomId, recv, conn.ID)
}

func (s *ChatService) handleInviteRoom(conn *Connection, env *chatv1.ChatEnvelope) {
	if conn.UserID == "" {
		s.sendError(conn, "UNAUTHORIZED", "not authenticated")
		return
	}
	req := env.GetInviteRoomRequest()
	if req == nil {
		s.sendError(conn, "BAD_REQUEST", "missing invite room request")
		return
	}

	// Verify inviter is in the room
	roomUID, err := uuid.Parse(req.RoomId)
	if err != nil {
		s.sendError(conn, "BAD_REQUEST", "invalid room id")
		return
	}
	members, err := s.queries.GetRoomMembers(context.Background(), roomUID)
	if err != nil {
		s.sendError(conn, "NOT_FOUND", "room not found")
		return
	}
	inRoom := false
	for _, m := range members {
		if m.UserID.String() == conn.UserID {
			inRoom = true
			break
		}
	}
	if !inRoom {
		s.sendError(conn, "FORBIDDEN", "you are not in this room")
		return
	}

	// Forward invite to target user
	notify := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_INVITE_ROOM,
		Payload: &chatv1.ChatEnvelope_InviteRoomNotify{
			InviteRoomNotify: &chatv1.InviteRoomNotify{
				RoomId:         req.RoomId,
				InviterId:      conn.UserID,
				InviterUsername: "",
			},
		},
	}
	s.server.SendToUser(req.TargetUserId, notify)
}

func (s *ChatService) sendEnvelope(conn *Connection, env *chatv1.ChatEnvelope) {
	data, _ := proto.Marshal(env)
	conn.Send(Frame{Type: uint16(env.Type), Value: data})
}

func (s *ChatService) broadcastEnvelope(roomID string, env *chatv1.ChatEnvelope, exclude string) {
	data, _ := proto.Marshal(env)
	s.server.Broadcast(roomID, Frame{Type: uint16(env.Type), Value: data}, exclude)
}

func (s *ChatService) sendError(conn *Connection, code, message string) {
	resp := &chatv1.ChatEnvelope{
		Type: chatv1.MessageType_ERROR,
		Payload: &chatv1.ChatEnvelope_Error{
			Error: &chatv1.ErrorMessage{Code: code, Message: message},
		},
	}
	s.sendEnvelope(conn, resp)
}
