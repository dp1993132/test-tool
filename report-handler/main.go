package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main(){
	defer recover()
	fl,_:=os.OpenFile("report-logs",os.O_CREATE|os.O_WRONLY,0644)
	defer fl.Close()
	log.SetOutput(fl)
	http.HandleFunc("/", func(resp http.ResponseWriter, request *http.Request) {
		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf("%s ok",time.Now().String())))

		log.Println(request.URL.Path, "ok")
	})
	http.ListenAndServe(":8081",nil)
}
