package main

import (
	"github.com/yunhanshu-net/runcher/kernel"
	_ "net/http/pprof"
)

func main() {

	app := kernel.NewRuncher()
	err2 := app.Run()
	if err2 != nil {
		panic(err2)
	}

	select {}
}
