package coder

import (
	"github.com/yunhanshu-net/pkg/x/gitx"
	"github.com/yunhanshu-net/pkg/x/jsonx"
)

type GitCommitMsg struct {
	Version string `json:"version"`
	Msg     string `json:"msg"`
}

func (m *GitCommitMsg) JSON() string {
	return jsonx.String(m)
}

func InitGit(coder *GoCoder) (*gitx.GitProject, error) {
	open, err := gitx.InitOrOpen(coder.CodePath, coder.User, coder.User+"@yunhanshu.net")
	if err != nil {
		return nil, err
	}
	return open, nil
}
