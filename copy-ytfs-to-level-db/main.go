package main

import (
	"fmt"
	"github.com/dp1993132/test-tool/stress-write-ytfs/config"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/yottachain/YTFS/storage"
	"log"
	"sync/atomic"
	"time"
	"path"
	"github.com/yottachain/YTDataNode/util"
)


var num uint64
var preNum uint64
var cfg *config.Config
var end chan struct{}
var ti *storage.TableIterator
var preTime time.Time

var ldb *leveldb.DB

func init(){
	preTime = time.Now()
	end = make(chan struct{})
	cfg,_=config.ReadConfig()
	ti,_=storage.GetTableIterator(path.Join(util.GetYTFSPath(),"index.db"),cfg.Options)
	db,err := leveldb.OpenFile(path.Join(util.GetYTFSPath(),"maindb"),nil)
	if err != nil {
		panic(err)
	}
	ldb = db
}

func main(){
	go printSpeed()
	for {
		tb,err:=ti.GetNoNilTableBytes()
		if err !=nil {
			break
		}
		lbatch := new(leveldb.Batch)
		for k,v:=range tb{
			lbatch.Put(k[:],v)
		}
		err=ldb.Write(lbatch,nil)
		if err != nil {
			log.Println(err)
		}
		atomic.AddUint64(&num,uint64(len(tb)))
	}
	end<- struct{}{}
}

func printSpeed () {
	for {
		select {
		case <-end:
			return
		default:
			<-time.After(time.Second)
			addCount:=atomic.LoadUint64(&num)-preNum
			if addCount == 0 {
				continue
			}
			log.Printf("copy %d 耗时 %f s进度 %d/%d 百分比 %.4f 预计完成剩余时间%s\n",
				addCount,
				time.Now().Sub(preTime).Seconds(),
				num,ti.Len(),
				float64(num)/(float64(ti.Len())+1)*100,
				FormatTd((ti.Len()-num)/addCount),
			)
			preNum = atomic.LoadUint64(&num)
			preTime = time.Now()
		}
	}
}

func FormatTd(td uint64) string {
	const (
		m = 60
		h = 60 * m
		d = 60 * 24
		)

	if td > d {
		return fmt.Sprintf("%d 天",td/d)
	}
	if td > h {
		return fmt.Sprintf("%d 时",td/h)
	}
	if td > m {
		return fmt.Sprintf("%d 分",td/m)
	}
	return fmt.Sprintf("%d 秒",td)
}
