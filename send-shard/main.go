package main

import (
	"context"
	//"crypto/md5"
	"fmt"
	"github.com/dp1993132/test-tool/send-shard/message"
	"github.com/dp1993132/test-tool/send-shard/token"
	"github.com/golang/protobuf/proto"
	"github.com/graydream/YTHost/client"
	"github.com/graydream/YTHost/option"
	"github.com/graydream/YTHost"
	. "github.com/graydream/YTHost/hostInterface"
	"github.com/libp2p/go-libp2p-core/peer"
	//"github.com/mr-tron/base58"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
	_ "net/http/pprof"
	"net/http"
)
var r0 = rand.New(rand.NewSource(time.Now().UnixNano()))
var hst  Host
var l sync.RWMutex
var sleepTime = 0
var listenAddr = "/ip4/0.0.0.0/tcp/8999"

var addrs = []string{
	"/ip4/127.0.0.1/tcp/9001/p2p/16Uiu2HAkuTUCLMEvN4UeCSsYLPTbdcWLoyMtjPhY9DaQ6CdCUwoU",
}
var count int = 0
var concurrency = 10

var successTime chan struct{}
var errorTime chan struct{}

var queue chan struct{}
var timechange chan struct{}
var needToken = false

func main(){
	go func() {
		http.ListenAndServe(":10001",nil)
	}()
	runTest()
}
func runTest(){

	startTime := time.Now()
	wg := sync.WaitGroup{}
	fl,err := os.OpenFile(os.Args[1],os.O_RDONLY,0666)
	if err != nil {
		fmt.Println(err)
	}
	viper.SetConfigType("yaml")
	err =viper.ReadConfig(fl)
	if err != nil {
		fmt.Println(err)
	}

	addrs = viper.GetStringSlice("addrs")
	count = viper.GetInt("count")
	concurrency = viper.GetInt("concurrency")
	sleepTime = viper.GetInt("sleep")
	needToken = viper.GetBool("needToken")
	listenAddr = viper.GetString("listenAddr")
	queue = make(chan struct{},concurrency)

	successTime = make(chan struct{}, len(addrs)*count)
	errorTime = make(chan struct{}, len(addrs)*count)
	timechange = make(chan struct{}, len(addrs)*count)
	wg.Add(len(addrs)*count)
	fmt.Println(addrs,count, concurrency)


	ma,_:=multiaddr.NewMultiaddr(listenAddr)
	h,err := host.NewHost(option.ListenAddr(ma))
	log.Println("监听：",listenAddr)
	if err != nil {
		fmt.Println(err)
	}
	hst = h
	for i:=0;i<count;i++{
		go func() {
			for _,v:= range addrs{
				ctx,_ := context.WithTimeout(context.Background(),60 * time.Second)
				ma,err:= multiaddr.NewMultiaddr(v)
				if err!=nil{
					fmt.Println(err)
					continue
				}
				info,err := peer.AddrInfoFromP2pAddr(ma)
				if err!=nil{
					fmt.Println(err)
					continue
				}
				startT:=time.Now()
				clt,err := hst.ClientStore().Get(ctx,info.ID,info.Addrs)
				endT:=time.Now()
				fmt.Println("connect use",endT.Sub(startT).Seconds())
				//defer clt.Close()
				if err!=nil{
					fmt.Println(err)
					continue
				}
				go func() {
					err=send(info.ID,clt)
					if err!=nil{
						fmt.Println(err)
						errorTime <- struct{}{}
					} else {
						successTime <- struct{}{}
					}
					defer func() {
						timechange <-struct {}{}
					}()
					defer wg.Done()
				}()
			}
		}()
	}
	wg.Wait()
	fmt.Printf("success[%d],error[%d], total:[%d],time-consuming[%f s]\n",len(successTime),len(errorTime),len(successTime)+len(errorTime),time.Now().Sub(startTime).Seconds())
	select {

	}
}

func getTK(id peer.ID,clt *client.YTHostClient) (string,error){
	if needToken == false {
		tk := token.NewToken()
		tk.Tm = time.Now()
		return tk.String(),nil
	}
	ctx,cancel := context.WithTimeout(context.Background(),1 * time.Second)
	defer cancel()
	var msg = message.MsgIDNodeCapacityRequest
	buf,err:=clt.SendMsg(ctx,message.MsgIDNodeCapacityRequest.Value(),msg.Bytes())
	var res message.NodeCapacityResponse
	err= proto.UnmarshalMerge(buf[2:],&res)
	if err != nil {
		return "",err
	}
	return res.AllocId,nil
}


func send(id peer.ID,clt *client.YTHostClient) error{

	l.Lock()
	//var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	l.Unlock()
	queue <- struct{}{}
	defer func() {<-queue}()
	//ctx,cancel := context.WithTimeout(context.Background(),1 * time.Second)
	//defer cancel()
	var tk token.Token
	tkstr,err := getTK(id,clt)
	err = tk.FillFromString(tkstr)
	if err != nil {
		return err
	}
	//var data []byte
	//data = make([]byte,16*1024)
	//r.Read(data)
	//m5 := md5.New()
	//m5.Reset()
	//m5.Write(data)
	//key := m5.Sum(nil)

	//var msg message.UploadShardRequest
	//msg.DAT = data
	//msg.VHF = key
	//msg.VBI = 1
	//msg.AllocId = tk.String()
	//buf,err:=proto.Marshal(&msg)
	//if err != nil {
	//	return err
	//}
	//
	//res,err:=clt.SendMsg(ctx,message.MsgIDUploadShardRequest.Value(),buf)
	//if err != nil {
	//	return err
	//}
	//
	//var resMsg message.UploadShard2CResponse
	//err = proto.Unmarshal(res[2:], &resMsg)
	//if err != nil {
	//	return err
	//}
	//if resMsg.RES != 0{
	//	return fmt.Errorf("错误：%d",resMsg.RES)
	//}
	//fmt.Printf("send shard[%s] to [%s] success\n",base58.Encode(msg.VHF),id)
	return nil
}
