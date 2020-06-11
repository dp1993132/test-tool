package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/yottachain/YTDataNode/message"
	"github.com/yottachain/YTHost/hostInterface"
	"github.com/yottachain/YTHost/option"
	"github.com/multiformats/go-multiaddr"
	host "github.com/yottachain/YTHost"
	"os"
	"sync"
	"time"
)

var l string
var r string
var q uint
var t int

func main() {
	flag.StringVar(&l,"l","/ip4/0.0.0.0/tcp/9003", "本地监听地址")
	flag.StringVar(&r,"r","", "远程地址")
	flag.IntVar(&t,"t",5,"超时时间")
	flag.UintVar(&q,"q",10,"并发数量")
	flag.Parse()

	lma,err:=multiaddr.NewMultiaddr(l)
	if err != nil {
		fmt.Println(err.Error())
		return
	}


	hst,err := host.NewHost(option.ListenAddr(lma))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ppool:=make(chan struct{},q)
	sc := bufio.NewScanner(os.Stdin)
	wg:=sync.WaitGroup{}
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		ppool<- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				<-ppool
			}()
			td,err:=ping(hst,line)
			if err != nil {
				os.Stdout.WriteString(fmt.Sprintf("%s,%s,%d\n", line,err.Error(),td.Milliseconds()))
			} else {
				os.Stdout.WriteString(fmt.Sprintf("%s,%s,%d\n", line,"ok",td.Milliseconds()))
			}
		}()
	}

	wg.Wait()
}

func ping(hst hostInterface.Host,r string)(time.Duration,error){

	rma,err:=multiaddr.NewMultiaddr(r)
	if err != nil {
		return 0,fmt.Errorf(err.Error())
	}
	pi,err:=peer.AddrInfoFromP2pAddr(rma)
	if err != nil {
		return 0,fmt.Errorf(err.Error())
	}
	ctx,cancel:=context.WithTimeout(context.Background(),time.Duration(t) * time.Second)
	defer cancel()

	timeS:=time.Now()
	clt,err:=hst.ClientStore().Get(ctx,pi.ID,pi.Addrs)
	if err != nil {
		return time.Now().Sub(timeS),fmt.Errorf(err.Error())
	}

	var msg message.NodeCapacityRequest

	buf,err := proto.Marshal(&msg)
	if err != nil {
		return time.Now().Sub(timeS),fmt.Errorf(err.Error())
	}

	_,err=clt.SendMsg(ctx,message.MsgIDString.Value(),buf)
	if err != nil {
		return time.Now().Sub(timeS),fmt.Errorf(err.Error())
	}

	return time.Now().Sub(timeS),nil
}
