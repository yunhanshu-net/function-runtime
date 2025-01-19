package service

import "github.com/yunhanshu-net/runcher/kernel"

type Service struct {
	Runcher *kernel.Runcher
}

func NewService() *Service {
	return &Service{
		Runcher: kernel.NewDefaultRuncher(),
	}
}
