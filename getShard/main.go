package main

import (
	"context"
	"flag"
	host "github.com/yottachain/YTHost"
)

var addr string
var vhf string

func main(){
	flag.StringVar(&addr,"a","","远程地址")
	flag.StringVar(&vhf,"h","","分片哈希")
	flag.Parse()

	h,_:=host.NewHost()
	h.Connect(context.Background(),)
}