//go:build linux || darwin || freebsd || openbsd || netbsd

package osx

import "golang.org/x/sys/unix"

func SyncFS() error {
	return unix.Sync() // 仅在 Unix 系统生效
}
