package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dp1993132/test-tool/test-tool/token"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/yottachain/YTHost/ClientManage"
	//"github.com/yottachain/YTHost/hostInterface"
	"hash"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dp1993132/test-tool/test-tool/message"
	"github.com/gogo/protobuf/proto"
	"github.com/multiformats/go-multiaddr"
	host "github.com/yottachain/YTHost"
	"github.com/yottachain/YTHost/option"
)

type Pi struct {
	ID    string `json:"nodeid"`
	Addrs string `json:"ip"`
}

type TestParams struct {
	Name string `json:"name"`
	RandNodeNum int `json:"randNodeNum"`
	OptimitionNodeNum int `json:"optimitionNodeNum"`
	QueueLength int `json:"queueLength"`
	TestDuration time.Duration `json:"testDuration"`
	CloseConn bool `json:"close_conn"`
	Timeout int `json:"timeout"`
}

var list []Pi
var count uint64
var dataOringin string
var qs int
var listenaddr string
var closeConn bool

var timeSum time.Duration
var maxId int64
var maxTime time.Duration = 0
var minId int64
var minTime time.Duration = 0
var timeout int

var HttpListenAddr string

var ids []peer.ID
var currentNodeList []peer.AddrInfo

var successCount uint32
var speed uint32

var optNumber = 0
var randNumber = 0
var updateNodeListInterval time.Duration = 10
var hst *ClientManage.Manager

var c chan os.Signal = make(chan os.Signal)

var taskList []TestParams
var taskFilePath string
var taskChan = make(chan TestParams,1000)

//var otzr = optimizer.New()

func main() {
	startTime := time.Now()
	defer func() {
		fmt.Println("总耗时",time.Now().Sub(startTime).Seconds(),"s")
	}()

	flag.StringVar(&dataOringin, "d", "/dev/urandom", "数据源")
	//flag.DurationVar(&duration,"duration",time.Second * 10,"执行时长")
	flag.StringVar(&listenaddr, "l", "/ip4/0.0.0.0/tcp/9003", "监听地址")
	flag.StringVar(&HttpListenAddr, "hl", "0.0.0.0:9004", "http服务器监听地址")
	flag.IntVar(&timeout, "ot", 10, "超时时间")
	flag.DurationVar(&updateNodeListInterval,"ui",time.Second * 10,"节点更新周期")
	flag.StringVar(&taskFilePath,"tf","","测试任务列表文件")

	flag.Parse()
	log.SetOutput(os.Stdout)

	getTargetNodeList()
	listenHost()

	go PrintSpeed(time.Second)
	go AutoGetNodeList()
	go HttpServer()

	go RunTest()

	signal.Notify(c,syscall.SIGINT,syscall.SIGQUIT,syscall.SIGHUP)
	<-c
}

func RunTest(){
	if taskFilePath!=""{
		fl,err:= os.OpenFile(taskFilePath,os.O_RDONLY,0644)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		dc:=json.NewDecoder(fl)
		dc.Decode(&taskList)
	}

	for _,v:=range taskList{
		taskChan <-v
	}

	for {
		task:=<-taskChan
		fmt.Printf("开始执行测试方案:%s 并发 %d 优化节点数 %d 随机节点数 %d\n",
			task.Name,
			task.QueueLength,
			task.OptimitionNodeNum,
			task.RandNodeNum,
		)
		logfile,_:=os.OpenFile(fmt.Sprintf("%s.log",task.Name),os.O_CREATE|os.O_TRUNC|os.O_WRONLY,0644)
		log.SetOutput(logfile)

		execTest(task)

		log.SetOutput(os.Stdout)
		fmt.Println(task.Name,"完成")
	}
}

// 获取待测试节点列表
func getTargetNodeList(){

	type NodeInfo struct {
		ID    string `json:"nodeid"`
		Addrs string `json:"ip"`
	}

	resp, err := http.Get("http://39.105.184.162:8082/active_nodes")
	if err != nil {
		println("error:获取节点列表失败")
		return
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if err != nil {
			println("error:获取节点列表失败")
			return
		}
		return
	}

	err = json.Unmarshal(buf, &list)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _,v:= range list{
		id,err := peer.Decode(v.ID)
		if err != nil {
			continue
		}
		ids = append(ids,id)
	}
}

func listenHost(){
	lsma, err := multiaddr.NewMultiaddr(listenaddr)
	if err != nil {
		log.Fatal(err)
	}
	h, err := host.NewHost(option.ListenAddr(lsma),option.OpenPProf(":10000"))
	if err != nil {
		log.Fatal(err)
	}
	//hst = h
	hst,err = ClientManage.NewManager(h.ClientStore())

	go hst.Keep(time.Second * 10)
}

// 执行测试方案
func execTest(params TestParams){
	defer func() {
		err:=recover()
		if err != nil {
			fmt.Println(err.(error).Error())
		}
	}()

	//otzr = optimizer.New()

	var queue = make(chan struct{}, params.QueueLength)
	randNumber = params.RandNodeNum
	optNumber = params.OptimitionNodeNum
	closeConn = params.CloseConn
	timeout = params.Timeout

	dataReader, err := os.OpenFile(dataOringin, os.O_RDONLY, 0644)
	defer dataReader.Close()
	if err != nil {
		log.Fatal(err)
	}
	var m5 = md5.New()

	<-time.After(time.Second)
	//go otzr.Run(context.Background())


	ctx ,_:=context.WithTimeout(context.Background(),time.Minute*params.TestDuration)
	for {
		<-time.After(time.Millisecond * 10)
		for _,p := range currentNodeList {
			select {
			case <-ctx.Done():
				return
			default:
				queue <- struct{}{}
				go TestFunc(queue,p, m5, dataReader)
			}
		}
	}
}

func TestFunc(queue chan struct{},p peer.AddrInfo, m5 hash.Hash, dataReader *os.File) {

	defer func() {
		recover()
	}()
	defer func() {
		<-queue
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()

	var data = make([]byte, 16*1024)
	dataReader.Read(data)
	m5.Reset()
	m5.Write(data)
	//var vhf = m5.Sum(nil)
	//var b58vhf = base58.Encode(vhf)
	//startTime := time.Now()
	const (
		send_ok = 0
		error_conn = 1
		error_token = 2
		error_send = 3
		error_other = 4
	)

	var opres = error_send

	defer func() {
		//otzr.Feedback(counter.InRow{p.ID.Pretty(),opres})
		log.Println(p.ID,"status:",opres)
	}()

	clt, err := hst.Get(p.ID,p.Addrs)
	if err != nil {
		opres = error_conn
		return
	}

	var tkMsg message.NodeCapacityRequest
	var tkResMsg message.NodeCapacityResponse

	tkMsgBuf,err := proto.Marshal(&tkMsg)
	if err != nil {
		opres = error_other
		return
	}

	tkResBuf,err:=clt.SendMsg(ctx,message.MsgIDNodeCapacityRequest.Value(),tkMsgBuf)
	if err != nil {
		opres = error_send
		return
	}

	proto.Unmarshal(tkResBuf[2:],&tkResMsg)

	if tkResMsg.AllocId == "" {
		//log.Println("获取token失败")
		opres = error_token
		return
	}  else {
		//log.Println("tk:",tkResMsg.AllocId)
	}

	token := token.NewToken()
	token.Tm = time.Now()

	var uploadReqMsg message.UploadShardRequestTest
	uploadReqMsg.DAT = data

	uploadReqData, err := proto.Marshal(&uploadReqMsg)
	if err != nil {
		opres = 4
		return
	}

	_, err = clt.SendMsg(ctx, message.MsgIDSleepReturn.Value(), uploadReqData)
	if err != nil {
		opres = 3
		return
	}
	// 成功计数
	atomic.AddUint32(&successCount,1)
	opres = send_ok

	//if closeConn {
	//	clt.Close()
	//}
}

func PrintSpeed(t time.Duration){
	for {
		atomic.StoreUint32(&successCount,0)
		<-time.After(t)
		sc:=atomic.LoadUint32(&successCount)
		speed = sc*16
		log.Printf("上传速度 %d KB\\s\n",speed)
	}
}

func GetIDS()[]string{
	var res = make([]string,len(list))
	for i,_ := range list{
		res[i]=list[i].ID
	}
	return res
}

func GetPi(id string) *Pi{
	for _,v:= range list {
		if v.ID == id {
			return &v
		}
	}
	return nil
}


func GetNodeList(optNumber int, randNumber int) []peer.AddrInfo{
	//sortList:=otzr.Get(ids...).Sort()
	//optList:=sortList[:optNumber]
	//var res=make([]*Pi,optNumber+randNumber)
	//
	//for k,v := range optList{
	//	res[k]=GetPi(v.ID)
	//}
	//
	//for i:=0;i<randNumber;i++ {
	//	index := rand.Intn(len(sortList)-optNumber-1)+optNumber
	//	if index >= len(res){
	//		continue
	//	}
	//	res[optNumber+i]=GetPi(sortList[optNumber:][index].ID)
	//}
	return hst.GetOptNodes(ids,optNumber+randNumber)
}

func AutoGetNodeList(){
	for {
		fmt.Println("节点列表更新")
		currentNodeList= GetNodeList(optNumber,randNumber)
		<-time.After(updateNodeListInterval)
	}
}

//func GetStatus()[]byte{
//	//buf:=bytes.NewBuffer([]byte{})
//
//	//fmt.Fprintf(buf,"%s,%s,%s,%s,%s,%s\n","ID","success","error_conn","error_token","error_send_msg","error_other")
//	//for k,v:=range otzr.CurrentCount(){
//	//
//	//	fmt.Fprintf(buf,"%s,%d,%d,%d,%d,%d\n",k,v[0],v[1],v[2],v[3],v[4])
//	//}
//	//return buf.Bytes()
//}

func HttpServer(){
	//http.HandleFunc("/status", func(writer http.ResponseWriter, request *http.Request) {
	//
	//})
	//http.HandleFunc("/speed", func(writer http.ResponseWriter, request *http.Request) {
	//	fmt.Fprintf(writer,"%d KB/s",speed)
	//})
	//http.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
	//
	//	var params TestParams
	//	dc:=json.NewDecoder(request.Body)
	//	dc.Decode(&params)
	//
	//	taskChan<-params
	//})
	//http.Handle("/",http.FileServer(http.Dir(path.Dir(os.Args[0]))))
	//http.ListenAndServe(HttpListenAddr,nil)
}

