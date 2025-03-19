//go:build windows

package osx

func SyncFS() error {
	// Windows 无 sync 系统调用，可留空或实现替代逻辑
	return nil
}
