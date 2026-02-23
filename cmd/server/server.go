// Package server 提供服务器启动功能
package server

import (
	"context"
	"log/slog"
)

// Server 服务器函数类型
type Server func(context.Context) error

// StartServer 启动服务器
func StartServer(ctx context.Context, servers ...Server) error {
	for _, srv := range servers {
		go func(s Server) {
			if err := s(ctx); err != nil {
				slog.Error("Failed to start server", "error", err)
			}
		}(srv)
	}
	return nil
}
