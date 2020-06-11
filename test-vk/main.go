package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"log"
	"time"
)

type Data struct {
	pr crypto.PrivKey
	pu crypto.PubKey
	data []byte
	sig []byte
}

func main(){

	var datas []Data
	var number int
	var _v bool
	flag.IntVar(&number, "n", 10000,"number")
	flag.BoolVar(&_v,"v",false,"v")
	flag.Parse()
	datas = make([]Data, number)
	var startTime = time.Now()
	for i:=0;i<number;i++{
		var data []byte
		data = make([]byte,16)
		pri,pub,err:=crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			log.Fatal(err)
		}
		rand.Read(data)
		sigData,err:=pri.Sign(data)
		if err != nil {
			log.Fatal(err)
		}
		datas[i]=Data{
			pri,
			pub,
			data,
			sigData,
		}

	}
	fmt.Printf("用时%f秒,签名%d个\n", time.Now().Sub(startTime).Seconds(),len(datas))
	startTime = time.Now()
	for _,v:= range datas {
		ok,err:=v.pu.Verify(v.data,v.sig)
		if _v{
			log.Println(ok,err)
		}

	}
	fmt.Printf("用时%f秒,验证%d个\n", time.Now().Sub(startTime).Seconds(),len(datas))
}

