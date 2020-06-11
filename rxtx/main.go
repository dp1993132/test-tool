package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main(){
	nf, err := os.OpenFile("/proc/net/dev", os.O_RDONLY, 0644)
	if err != nil {
		return
	}

	sc := bufio.NewScanner(nf)

	var rx,tx int64

	for sc.Scan() {
		line:=sc.Text()
		arr:=strings.Split(line,":")
		if len(arr)>1{
			reg:=regexp.MustCompile(" +")
			arr2:=reg.Split(arr[1],-1)
			r,_:=strconv.ParseInt(arr2[1],10,64)
			t,_:=strconv.ParseInt(arr2[9],10,64)
			rx=rx+r
			tx=tx+t
		}
	}
	fmt.Println("rx:",rx)
	fmt.Println("tx:",tx)
}