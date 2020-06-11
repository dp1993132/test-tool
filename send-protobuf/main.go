package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	//"github.com/yottachain/YTDataNode/message"
	"io/ioutil"
	"log"
	"os"
)

func main(){
	ma,err := multiaddr.NewMultiaddr(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	pi,err := peer.AddrInfoFromP2pAddr(ma)
	h,err:=libp2p.New(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	err=h.Connect(context.Background(),*pi)
	if err != nil {
		log.Fatal(err)
	}
	stm,err:=h.NewStream(context.Background(),pi.ID,"/node/0.0.2")
	if err != nil {
		log.Fatal(err)
	}
	fd,err:=os.OpenFile(os.Args[2],os.O_RDONLY,0644)
	if err != nil {
		log.Fatal(err)
	}
	buf,err:=ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal(err)
	}
	ee := gob.NewEncoder(stm)
	err=ee.Encode(buf)
	if err != nil {
		log.Fatal(err)
	}
	resbuf,err :=ioutil.ReadAll(stm)
	fmt.Println(resbuf,err)
}