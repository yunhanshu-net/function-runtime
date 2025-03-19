package main

import (
	v2 "github.com/yunhanshu-net/runcher/v2"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	runcher := v2.NewRuncher()
	err := runcher.Run()
	if err != nil {
		panic(err)
	}
	defer runcher.Close()
	go func() {
		// 启动一个 http server，注意 pprof 相关的 handler 已经自动注册过了
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	select {}
}
