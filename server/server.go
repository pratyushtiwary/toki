package server

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

type Server struct {
	port      int
	listener  net.Listener
	channels  map[int]chan []byte
	passkey   atomic.Value
	inDevMode bool
	nextID    int
	rwLock    sync.RWMutex
	wg        sync.WaitGroup
}

func (s *Server) GetPort() int {
	return s.port
}

func (s *Server) SetPasskey(passkey string) {
	s.passkey.Store(passkey)
}

func (s *Server) Listen() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))

	if err != nil {
		return err
	}

	s.listener = listener

	go s.listenLoop()

	return nil
}

func (s *Server) listenLoop() {
	for {
		conn, err := s.listener.Accept()

		if err != nil {
			return
		}

		s.rwLock.Lock()
		ch := make(chan []byte, 100)
		s.channels[s.nextID] = ch
		s.wg.Add(1)
		go s.writeLoop(conn, s.nextID, ch)
		s.nextID += 1
		s.rwLock.Unlock()
	}
}

func (s *Server) writeLoop(conn net.Conn, clientId int, ch chan []byte) {
	defer s.removeClient(conn, clientId, ch)

	if !s.inDevMode {
		handshakePass := make([]byte, 36)

		_, err := conn.Read(handshakePass)

		if err != nil {
			return
		}

		expectedKey, ok := s.passkey.Load().(string)

		if !ok {
			return
		}

		if string(handshakePass) != expectedKey {
			return
		}
	}

	for message := range ch {
		_, err := conn.Write(message)

		if err != nil {
			return
		}
	}
}

func (s *Server) removeClient(conn net.Conn, clientId int, ch chan []byte) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	if existingCh, exists := s.channels[clientId]; exists && existingCh == ch {
		close(ch)
		delete(s.channels, clientId)
	}
	conn.Close()
	s.wg.Done()
}

func (s *Server) Broadcast(message string) {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	messageByte := []byte(message + "\n")

	for _, ch := range s.channels {
		select {
		case ch <- messageByte:
		default:
		}
	}
}

func (s *Server) Close() {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	for _, ch := range s.channels {
		close(ch)
	}
	s.wg.Wait()
	s.listener.Close()
}

func NewServer(port int, inDevMode bool) *Server {
	return &Server{
		port:      port,
		channels:  make(map[int]chan []byte),
		inDevMode: inDevMode,
	}
}
