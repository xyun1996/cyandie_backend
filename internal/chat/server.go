package chat

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	chatv1 "github.com/cyandie/backend/api/proto/chat/v1"
	"google.golang.org/protobuf/proto"
)

type TCPServer struct {
	addr             string
	listener         net.Listener
	mu               sync.RWMutex
	conns            map[string]*Connection
	onConnect        func(conn *Connection)
	onDisconnect     func(conn *Connection)
	onMessage        func(conn *Connection, frame *Frame)
	authTimeout      time.Duration
	heartbeatTimeout time.Duration
}

type Connection struct {
	ID       string
	Conn     net.Conn
	UserID   string
	Username string
	rooms    map[string]struct{}
	mu       sync.Mutex
}

func (c *Connection) JoinRoom(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rooms == nil {
		c.rooms = make(map[string]struct{})
	}
	c.rooms[roomID] = struct{}{}
}

func (c *Connection) LeaveRoom(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rooms != nil {
		delete(c.rooms, roomID)
	}
}

func (c *Connection) IsInRoom(roomID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rooms == nil {
		return false
	}
	_, ok := c.rooms[roomID]
	return ok
}

func NewTCPServer(addr string, authTimeout, heartbeatTimeout time.Duration) *TCPServer {
	return &TCPServer{
		addr:             addr,
		conns:            make(map[string]*Connection),
		authTimeout:      authTimeout,
		heartbeatTimeout: heartbeatTimeout,
	}
}

func (s *TCPServer) OnConnect(fn func(conn *Connection))       { s.onConnect = fn }
func (s *TCPServer) OnDisconnect(fn func(conn *Connection))    { s.onDisconnect = fn }
func (s *TCPServer) OnMessage(fn func(conn *Connection, frame *Frame)) { s.onMessage = fn }

func (s *TCPServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln
	slog.Info("TCP server listening", "addr", s.addr)

	go s.acceptLoop()
	return nil
}

func (s *TCPServer) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *TCPServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		c := &Connection{ID: generateConnID(), Conn: conn}
		s.mu.Lock()
		s.conns[c.ID] = c
		s.mu.Unlock()

		if s.onConnect != nil {
			s.onConnect(c)
		}

		// Close connection if it doesn't authenticate within authTimeout.
		if s.authTimeout > 0 {
			go func() {
				time.Sleep(s.authTimeout)
				c.mu.Lock()
				authenticated := c.UserID != ""
				c.mu.Unlock()
				if !authenticated {
					c.Conn.Close()
				}
			}()
		}

		go s.readLoop(c)
	}
}

func (s *TCPServer) readLoop(c *Connection) {
	defer func() {
		c.Conn.Close()
		s.mu.Lock()
		delete(s.conns, c.ID)
		s.mu.Unlock()
		if s.onDisconnect != nil {
			s.onDisconnect(c)
		}
	}()

	if s.heartbeatTimeout > 0 {
		c.Conn.SetReadDeadline(time.Now().Add(s.heartbeatTimeout))
	}

	for {
		frame, err := DecodeFrame(c.Conn)
		if err != nil {
			return // includes deadline exceeded
		}
		if s.heartbeatTimeout > 0 {
			c.Conn.SetReadDeadline(time.Now().Add(s.heartbeatTimeout))
		}
		if s.onMessage != nil {
			s.onMessage(c, frame)
		}
	}
}

func (c *Connection) Send(frame Frame) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.Conn.Write(EncodeFrame(frame))
	return err
}

func (s *TCPServer) Broadcast(roomID string, frame Frame, exclude string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.conns {
		if c.ID != exclude && c.UserID != "" && c.IsInRoom(roomID) {
			c.Send(frame)
		}
	}
}

func (s *TCPServer) GetConnection(userID string) (*Connection, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.conns {
		if c.UserID == userID {
			return c, true
		}
	}
	return nil, false
}

// SendToUser sends a protobuf envelope to a specific user if they are connected.
func (s *TCPServer) SendToUser(userID string, msg *chatv1.ChatEnvelope) {
	conn, ok := s.GetConnection(userID)
	if !ok {
		return
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return
	}
	frame := Frame{Type: uint16(msg.Type), Value: data}
	conn.Send(frame)
}

func generateConnID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}
