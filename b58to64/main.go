package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/mr-tron/base58"
	"os"
)

func main(){
	scanner:=bufio.NewScanner(os.Stdin)
	for scanner.Scan(){
		line := scanner.Text()
		buf,err:=base64.StdEncoding.DecodeString(line)
		if err != nil {
			continue
		}
		out := base58.Encode(buf)
		fmt.Fprintln(os.Stdout,out)
	}
}
