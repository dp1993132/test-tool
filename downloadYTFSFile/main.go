package main

import (
	"context"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/yottachain/YTDataNode/message"
	host "github.com/yottachain/YTHost"
	"github.com/yottachain/YTHost/hostInterface"
	"github.com/yottachain/YTHost/option"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var selfIP string
var hst hostInterface.Host

const privKeyPem = `
-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQGRM+cWbPIqdnBsvC9OpjMMS2TRQ+0Uaqad6Sct8wlroumztjuu
u/hcJ49lcRm8IFAC7ioG2PsKMSd42klEQvoPCMlvC6lD58/H6p4rxo5SEajfySxC
oCq/PmIVj2R0lb3AIdWSmN391NIkzdrzB3RcIAv826ehS8W9wjSaJXOSeQIDAQAB
AoGAROHYZy1FVq0HGGQm6yX11cKMCLHU3pCBEtOY+najw2sxHL3I+XMkbQ4NvKjy
di5GDnq9lHdkgpd143X25lVpgQDzpmizyASmlbWR6Qle2Ld6UjDWrB4nkjnxyRfa
TTND2O8sj9cwD2y7I8pF1+YFlvsSJ4MiEATH2G9wlITXKJECQQDyhubKqFm3I/Tu
1QPORBhmZcHpiMkgsoJLkQLW5GxzqCJR2rlx7Oi78b4jTEzn5VLWPifLE8rjYMTA
do4Uk84nAkEBp32cFdEeyCMUUoXvMs86hrMbb3/eljpvsgUmhdLSxJjvAxdbBf/y
rcuMfhvukTEBFS09JXkZ3p1HGzNyJHKeXwJAaL7s1OBLBzcnZTNpFmAArdELJCLo
ww92CM8Ti95SHM2kLPgrmdG5Xtr0xOgCWzGHSnLD2wisWvIDaCCMEsUXhwJAOXcd
AAThdWz1LAGKpM1j9rVFKssiLCZ/05tJT+18tjq+bB/2NQk3KAgv50jpBYCt0e7S
lkwpi4CyDmnbukBnZwJASID2yAka1HcZkJGCljw3w/CriOYRBQ85EEFTAp8NaDXi
p8XF2yjvLBQtI1dTcBALYfbMC9kHQriVYWXTKtKpdg==
-----END RSA PRIVATE KEY-----
`

func main()  {
	Service()
}

func Sign()[]byte{
	m5 := md5.New()
	m5.Reset()
	m5.Write([]byte("yotta debug"))
	hash := m5.Sum(nil)

	pp, _ := pem.Decode([]byte(privKeyPem))

	privKey, err := x509.ParsePKCS1PrivateKey(pp.Bytes)
	if err != nil {
		return nil
	}
	sig,err:=rsa.SignPKCS1v15(rand.Reader,privKey,crypto.MD5,hash)
	if err != nil {
		return nil
	}
	return sig
}

func Service(){
	http.HandleFunc("/upload/", func(resp http.ResponseWriter, request *http.Request) {
		filename:=strings.ReplaceAll(request.URL.Path,"/upload",path.Dir(os.Args[0]))
		fi,err:=os.OpenFile(filename,os.O_CREATE|os.O_WRONLY|os.O_TRUNC,0644)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer fi.Close()
		io.Copy(fi,request.Body)
		defer request.Body.Close()
		defer log.Println("下载",filename,"成功")
	})
	
	http.HandleFunc("/download", func(resp http.ResponseWriter, request *http.Request) {
		body,err:=ioutil.ReadAll(request.Body)
		if err !=nil {
			resp.WriteHeader(500)
			resp.Write([]byte(err.Error()))
		}
		err=download(string(body))
		if err !=nil {
			resp.WriteHeader(500)
			resp.Write([]byte(err.Error()))
		}
		resp.Write([]byte("ok"))
	})
	log.Println("服务启动成功")
	http.ListenAndServe("0.0.0.0:9012",nil)
}

func download(addr string) error{
	fmt.Println("发送下载请求",addr)
	var msg message.DownloadYTFSFile
	msg.Gzip=true
	msg.Name="index.db"
	msg.Sig = Sign()
	msg.ServerUrl = selfIP

	buf,err:=proto.Marshal(&msg)
	if err != nil {
		return err
	}

	ma,err:=multiaddr.NewMultiaddr(addr)
	if err != nil {
		return err
	}
	pi,err:= peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		return err
	}

	ctx,cancel:= context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()
	clt,err:=hst.ClientStore().Get(ctx,pi.ID,pi.Addrs)
	if err != nil {
		return err
	}

	_,err=clt.SendMsg(context.Background(),message.MsgIDDownloadYTFSFile.Value(),buf)
	return err
}

func init(){
	laddr,_:=multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/9010")

	h,err := host.NewHost(option.ListenAddr(laddr))
	if err != nil {
		log.Fatal(err)
	}

	hst = h

	resp,err:=http.Get("http://123.57.81.177/self-ip")
	if err != nil {
		log.Fatal(err)
	} else {
		urlBuf,err:=ioutil.ReadAll(resp.Body)
		if err == nil{
			defer resp.Body.Close()
			selfIP = string(urlBuf)+":9012"
		}
	}
}

