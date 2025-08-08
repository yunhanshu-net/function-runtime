package coder

import (
	"github.com/yunhanshu-net/function-go/pkg/dto/api"
)

type RebuildProjectReq struct {
	Body    string `json:"body"`
	Name    string `json:"name"`    //应用名称（英文标识）
	Version string `json:"version"` //应用版本
	User    string `json:"user"`    //所属租户
}

type RebuildProjectResp struct {
	Hash           string      `json:"hash"`
	Apis           []*api.Info `json:"apis"`            //全部的api
	CurrentVersion string      `json:"current_version"` //当前版本
}
