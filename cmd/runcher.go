package cmd

import "github.com/yunhanshu-net/runcher/kernel"

var Runcher *kernel.Runcher

func Init() {
	Runcher = kernel.MustNewRuncher()
	Runcher.Run()
}
