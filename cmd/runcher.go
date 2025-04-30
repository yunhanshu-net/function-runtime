package cmd

import (
	"github.com/yunhanshu-net/runcher/kernel"
	"github.com/yunhanshu-net/runcher/service/http2nats"
)

var Runcher *kernel.Runcher

func Init() {
	Runcher = kernel.MustNewRuncher()
	err := Runcher.Run()
	if err != nil {
		panic(err)
	}

	http2nats.Setup(Runcher)

	//Runcher.Run()
}
