package actions

import "sync"

// 简单的内存注册表：key 为 "owner/name"，value 为 Action 实例
var (
	regMu sync.RWMutex
	reg   = map[string]Action{}
)

// Register 注册一个可用的 Action（幂等覆盖）
func Register(name string, a Action) {
	regMu.Lock()
	reg[name] = a
	regMu.Unlock()
}

// Get 获取已注册的 Action
func Get(name string) (Action, bool) {
	regMu.RLock()
	a, ok := reg[name]
	regMu.RUnlock()
	return a, ok
}
