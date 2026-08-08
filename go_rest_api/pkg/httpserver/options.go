package httpserver

import (
	"time"
)

// Option -.
type Option func(*Server)

// Addr -.
func Addr(addr string) Option {
	return func(s *Server) {
		s.address = addr
	}
}

// Prefork -.
func Prefork(prefork bool) Option {
	return func(s *Server) {
		s.prefork = prefork
	}
}

// ReadTimeout -.
func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = timeout
	}
}

// WriteTimeout -.
func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = timeout
	}
}

// ShutdownTimeout -.
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}
