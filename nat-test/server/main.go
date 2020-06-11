package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
)
var listenAddr *net.UDPAddr
var connAddr *net.UDPAddr

func main () {
	var la,ca string
	flag.StringVar(&la, "l","","监听地址")
	flag.StringVar(&ca, "c","","连接地址")
	flag.Parse()
	if la != "" {
		addr,err:=net.ResolveUDPAddr("udp", la)
		if err != nil {
			log.Fatal(err)
		}
		listenAddr = addr
		go listen()
	}
	if ca != "" {
		addr,err:=net.ResolveUDPAddr("udp", ca)
		if err != nil {
			log.Fatal(err)
		}
		connAddr = addr
		go connect()
	}
	select {
	}
}

func listen(){

	if conn,err:=net.ListenUDP("udp",listenAddr);err != nil{

	} else {
		log.Println("监听:",conn.LocalAddr().String())
		printConn(conn)
	}
}

func connect(){
	conn,err:=net.DialUDP("udp", listenAddr,connAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("连接成功：",conn.RemoteAddr().String(),conn.LocalAddr().String())
	go printConn(conn)
	for {
		var line string
		fmt.Scanln(&line)
		fmt.Fprintln(conn,line)
	}
}

func printConn(conn net.Conn){
	scanner:=bufio.NewScanner(conn)
	for scanner.Scan(){
		log.Println(scanner.Text())
	}
}
