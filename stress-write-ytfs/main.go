package main

import (
	"crypto/md5"
	"flag"
	"github.com/dp1993132/test-tool/stress-write-ytfs/config"
	ytfs "github.com/yottachain/YTFS"
	"github.com/yottachain/YTFS/common"
	"log"
	"math/rand"
	"os"
	util "yottachain/ytfs-util"
)

var count uint64

func main(){
	flag.Uint64Var(&count,"c",1,"写入次数")
	flag.Parse()

	fd,err := os.OpenFile("/dev/urandom",os.O_RDONLY,0644)
	if err != nil {
		log.Println(err)
		return
	}
	cfg,err:=config.ReadConfig()
	if err != nil {
		log.Println(err)
		return
	}
	yt,err := ytfs.Open(util.GetYTFSPath(),cfg.Options)
	if err != nil {
		log.Println(err)
		return
	}

	m5:=md5.New()

	var i uint64
	for ;i<count; {
		datas := make(map[common.IndexTableKey][]byte,10)
		c := rand.Intn(10)

		if c<=0 {
			continue
		}

		for j:=0;j<c;j++{

			data:=make([]byte,16*1024)
			fd.Read(data)
			m5.Reset()
			m5.Write(data)
			vhf:= m5.Sum(nil)
			var ha [16]byte
			copy(ha[:],vhf)
			datas[ha]=data
		}

		_,err=yt.BatchPut(datas)
		//for k,v:=range datas{
		//	yt.Put(k,v)
		//}
		if err != nil {
			log.Printf("[write] error %d-%d", i,c)
		} else {
			log.Printf("[write] success %d-%d", i,c)
		}

		i = i + uint64(c)
	}
}
