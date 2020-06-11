package main

import (
	"fmt"
	"net"
	"sync"
)

func main(){
	res := make(chan string,256)
	wg:=sync.WaitGroup{}
	wg.Add(256)
	go func() {
		for i:=0;i<256;i++{
			go func(i int){
				ok:=conn(i)
				if ok {
					res <- fmt.Sprintf("192.168.3.%d",i)
				} else {
					fmt.Printf("192.168.3.%d：连接失败\n",i)
				}
				defer wg.Done()
			}(i)
		}
	}()
	go func() {
		for {
			addr:=<-res
			fmt.Println("可以建立tcp连接:",addr)
		}
	}()
	wg.Wait()
}
func conn(end int) bool{
	_,err:=net.Dial("tcp",fmt.Sprintf("192.168.3.%d:22",end))
	return err == nil
}
