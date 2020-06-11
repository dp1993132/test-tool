package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dp1993132/test-tool/stress-write-ytfs/config"
	ytfs "github.com/yottachain/YTFS"
	"github.com/yottachain/YTFS/storage"
	"io/ioutil"
	"os"
	"path"
)

var src *ytfs.YTFS
var dist *ytfs.YTFS

var srcPath string
var distPath string

var srcConfig config.Config
var distConfig config.Config

var newSize uint64

const GB uint64 = 1 << 30
const TB uint64 = 1024 * GB

func main()  {
	flag.Uint64Var(&newSize,"new-size", 8,"新的空间大小,单位TB")
	flag.Parse()

	if len(os.Args) < 3 {
		fmt.Println("请传入正确的参数 command src dist")
		return
	}
	srcPath = os.Args[1]
	distPath = os.Args[2]

	srcConfigFd,err:=os.OpenFile(path.Join(srcPath,"config.json"),os.O_RDONLY,0644)
	if err != nil {
		fmt.Println("打开配置文件失败")
		return
	}
	srcConfigBuf,err:=ioutil.ReadAll(srcConfigFd)
	if err != nil {
		fmt.Println("读取配置文件失败")
		return
	}
	srcConfigFd.Close()
	err=json.Unmarshal(srcConfigBuf,&srcConfig)
	if err != nil {
		fmt.Println("配置文件解析失败")
		return
	}

	// 配置目标配置文件
	distConfig = srcConfig

	distConfig.Options=config.GetYTFSOptionsByParams(newSize*TB,1<<20)

	src,err=ytfs.Open(srcPath,srcConfig.Options)
	if err != nil {
		fmt.Println("打开源YTFS DB失败")
		return
	}
	dist,err=ytfs.Open(distPath,distConfig.Options)
	if err!=nil{
		fmt.Println("打开目标YTFS DB失败")
		return
	}

	it,err:=storage.GetTableIterator(path.Join(srcPath,"index.db"),srcConfig.Options)
	if err!=nil{
		fmt.Println("打开源YTFS DB失败",err.Error())
		return
	}

	var i uint64
	for {
		item,err:=it.GetNoNilTable()
		if err != nil && err.Error()=="table end"{
			break
		}
		for k,v := range item {
			i++
			if err:=dist.DB().Put(k,v);err != nil {
				fmt.Println("写入重复",i,err.Error())
				continue
			}
			fmt.Println("复制",i,"/",src.Len())
		}
	}

	fmt.Println("复制完成")

	distFd,err:=os.OpenFile(path.Join(distPath,"config.json"),os.O_WRONLY|os.O_CREATE|os.O_TRUNC,0644)
	if err != nil {
		fmt.Println("新配置创建失败")
		return
	}
	distbuf,err:=json.MarshalIndent(distConfig, "", "  ")
	if err != nil {
		fmt.Println("新配置编码失败")
		return
	}
	_,err=distFd.Write(distbuf)
	if err != nil {
		fmt.Println("新配置写入失败")
		return
	}
	distFd.Close()
}
