package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var display bool
var beginTime = time.Now()

func main () {
	var count uint64
	var queueSize uint
	var exec string
	var successNum uint64


	var taskChan chan byte

	flag.Uint64Var(&count, "c", 1, "循环次数")
	flag.UintVar(&queueSize, "qs", 1, "循环队列长度")
	flag.StringVar(&exec, "e", "", "执行的测试命令")
	flag.BoolVar(&display, "v", false, "是否显示执行结果")
	flag.Parse()

	taskChan = make(chan byte, queueSize)

	wg := sync.WaitGroup{}
	var i uint64
	t1 := time.Now()
	for i=0; i < count; i++{
		taskChan <- 1
		wg.Add(1)
		go func () {
			err := work(exec, i)
			if err == nil {
				successNum ++
			} else {
					fmt.Println(err)
			}
			<- taskChan
			wg.Done()
		}()
	}
	wg.Wait()
	t2 := time.Now()
	td := t2.Sub(t1)

	fmt.Printf("开始于%s 执行%d次 并发%d 耗时%f秒, 成功%d \n", beginTime.Format("06年1月2日 15:04:05.000"),count,queueSize,td.Seconds(), successNum)
}

func work(script string, n uint64) error {

	scriptArr := strings.Split(script, " ")
	c := exec.Command(scriptArr[0],scriptArr[1:]...)
	os.Setenv("P_INDEX", fmt.Sprintf("%d",n))
	c.Env = os.Environ()
	if display  {
		c.Stdout = os.Stdout
	}
	return c.Run()
}
