package conns

import (
	"github.com/yunhanshu-net/runcher/kernel"
	"github.com/yunhanshu-net/runcher/service"
)

var Runcher *kernel.Runcher
var RunnerHttp *service.RunnerHttp

func Init() {
	Runcher = kernel.NewRuncher()
	RunnerHttp = &service.RunnerHttp{
		R: Runcher,
	}

}
