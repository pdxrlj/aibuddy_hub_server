// Package websocket 提供 websocket 连接管理
package websocket

import (
	"sync"

	"github.com/olahol/melody"
)

// ConnPool 连接池
type ConnPool struct {
	*sync.Map
}

// NewConnPool 创建连接池
func NewConnPool() *ConnPool {
	return &ConnPool{
		Map: &sync.Map{},
	}
}

// Add 添加连接
func (p *ConnPool) Add(key string, session *melody.Session) {
	if p.Exists(key) {
		p.Close(key)
	}
	p.Store(key, session)
}

// Get 获取连接
func (p *ConnPool) Get(key string) (*melody.Session, bool) {
	session, ok := p.Load(key)
	if !ok {
		return nil, false
	}
	return session.(*melody.Session), true
}

// Remove 删除连接
func (p *ConnPool) Remove(key string) {
	p.Delete(key)
}

// Close 关闭连接
func (p *ConnPool) Close(key string) {
	session, ok := p.Get(key)
	if !ok {
		return
	}
	_ = session.Close()
	p.Remove(key)
}

// CloseAll 关闭所有连接
func (p *ConnPool) CloseAll() {
	p.Range(func(_, value any) bool {
		session := value.(*melody.Session)
		_ = session.Close()
		return true
	})
	p.Clear()
}

// Exists 检查连接是否存在
func (p *ConnPool) Exists(key string) bool {
	_, exists := p.Get(key)
	return exists
}
