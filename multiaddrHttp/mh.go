package main

import (
	"github.com/multiformats/go-multiaddr"
	"net/http"
	"net/rpc"
	"syscall"
)

func main() {
	ma,_:=multiaddr.NewMultiaddr("/dns4/www.baidu.com/tcp/80")
	s:=rpc.NewServer()
	syscall.Sendmsg()
}