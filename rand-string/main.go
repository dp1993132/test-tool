package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"syscall"
)



func main() {
	var l uint
	var buf = []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	r := rand.New(rand.NewSource(int64(syscall.Getegid())))
	fmt.Println(os.Getpid())
	flag.UintVar(&l, "l", 1, "字符串长度")
	flag.Parse()
	res := make([]byte, l)
	for k,_ := range res {
		res[k] =buf[r.Intn(len(buf))]
	}
	fmt.Printf("%s", res)
}
