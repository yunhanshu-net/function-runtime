package cmd

import (
	"github.com/yunhanshu-net/function-runtime/kernel"
	"github.com/yunhanshu-net/function-runtime/pkg/http2nats"
)

var Runcher *kernel.Runcher

func Init() {
	Runcher = kernel.MustNewRuncher()
	err := Runcher.Run()
	if err != nil {
		panic(err)
	}

	http2nats.Setup(Runcher.GetNatsConn())

	//Runcher.Run()
}
