package main

import (
	"bufio"
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/mr-tron/base58"
	"github.com/yottachain/YTDataNode/config"
	"github.com/yottachain/YTDataNode/util"
	ytfs "github.com/yottachain/YTFS"
	"github.com/yottachain/YTFS/common"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var ytpath = util.GetYTFSPath()

var outPath string
var listPath string

func main() {
	flag.StringVar(&outPath, "o", "/dev/stdout", "输出文件目录")
	flag.StringVar(&listPath, "l", "/dev/stdin", "待检查分片")
	flag.Parse()

	fd, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	infd,err:=os.OpenFile(listPath,os.O_RDONLY,0644)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(infd)

	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	yt,err:=ytfs.Open(util.GetYTFSPath(),cfg.Options)
	if err != nil {
		log.Fatal(err)
	}
	m5:=md5.New()
	fmt.Fprintf(fd,"vhf,dataMd5,checkStatus\n")
	for scanner.Scan(){
		line := scanner.Text()
		line = strings.Replace(line,"\n","",-1)
		vhf,err:=base58.Decode(line)
		if err != nil {
			fmt.Fprintf(fd,"%s,%s,%s\n",line,"",err.Error())
			continue
		}
		var vhfbuf [16]byte
		copy(vhfbuf[:],vhf)
		res,err:=yt.Get(common.IndexTableKey(vhfbuf))
		m5.Reset()
		m5.Write(res)
		resmd5:=m5.Sum(nil)
		if err != nil {
			fmt.Fprintf(fd,"%s,%s,%s\n",line,base58.Encode(resmd5),err.Error())
		} else if base58.Encode(resmd5)==line {
			fmt.Fprintf(fd,"%s,%s,%s\n",line,base58.Encode(resmd5),"success")
		} else {
			fmt.Fprintf(fd,"%s,%s,%s\n",line,base58.Encode(resmd5),"error")
		}
	}
}

func cls() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd.exe", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}
