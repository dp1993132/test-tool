package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	b = 1
	k = 1024 * b
	m = 1024 * k
	g = 1024 * m
)

var r0 = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	var bsstr string
	var countstr string
	var outfile string
	var bs uint64
	var count uint64
	var randAddCount uint64

	flag.StringVar(&bsstr, "bs", "16", "文件block大小，可以使用单位:b,k,m,g")
	flag.StringVar(&countstr, "count", "10", "文件写入次数")
	flag.StringVar(&outfile, "out", "/dev/stdout", "输出文件名，可以是设备")
	flag.Uint64Var(&randAddCount,"rand-add",0,"随机添加文件写入次数")
	flag.Parse()
	bs, err := ParseSize(bsstr)
	if err != nil {
		fmt.Println("参数错误：bs")
	}
	count, err = ParseSize(countstr)
	if randAddCount != 0{
		count = r0.Uint64() % randAddCount + count
	}
	if err != nil {
		fmt.Println("参数错误：count")
	}
	var i uint64 = 0
	wg := sync.WaitGroup{}
	wg.Add(int(count))

	file, err := os.OpenFile(outfile, os.O_CREATE|os.O_WRONLY, 0644)
	time1 := time.Now()
	defer file.Close()
	for ; i < count; i++ {
		go writeBlock(bs, file, &wg)
	}
	wg.Wait()
	time2 := time.Now()
	timed := time2.Sub(time1)
	fmt.Println()
	fmt.Println("耗时", timed.String(), outfile, "文件大小", bs*count/k, "k", float64(bs*count)/timed.Seconds()/m, "m/s")
}

func ParseSize(str string) (uint64, error) {
	var unit = str[len(str)-1:]
	var size uint64 = 0
	var err error

	switch unit {
	case "b", "B", "k", "K", "m", "M", "g", "G":
		size, err = strconv.ParseUint(str[:len(str)-1], 10, 64)
		switch unit {
		case "k", "K":
			size = size * k
		case "m", "M":
			size = size * m
		case "g", "G":
			size = size * g
		}
	default:
		return strconv.ParseUint(str, 10, 64)
	}
	return size, err
}

func writeBlock(bs uint64, out io.Writer, wg *sync.WaitGroup) {
	r := rand.New(rand.NewSource(r0.Int63()))
	var buf []byte
	var cbSize uint64 = 64 * k
	buf = make([]byte, cbSize)

	var i uint64
	for ; i < bs/cbSize; i++ {
		r.Read(buf)
		out.Write(buf)
	}
	if obSize := bs % cbSize; obSize > 0 {
		r.Read(buf)
		out.Write(buf[:obSize])
	}
	defer wg.Done()
}
