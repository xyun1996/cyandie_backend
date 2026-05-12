package chat

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

type TCPServer struct {
	addr        string
	listener    net.Listener
	mu          sync.RWMutex
	conns       map[string]*Connection
	onConnect   func(conn *Connection)
	onDisconnect func(conn *Connection)
	onMessage   func(conn *Connection, frame *Frame)
}

type Connection struct {
	ID     string
	Conn   net.Conn
	UserID string
	mu     sync.Mutex
}

func NewTCPServer(addr string) *TCPServer {
	return &TCPServer{
		addr:  addr,
		conns: make(map[string]*Connection),
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

	for {
		frame, err := DecodeFrame(c.Conn)
		if err != nil {
			return
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
		if c.ID != exclude && c.UserID != "" {
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

func generateConnID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}
