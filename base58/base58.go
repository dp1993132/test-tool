package main

import (
	"bufio"
	"flag"
	"github.com/mr-tron/base58"
	"os"
)

var isDecode bool = false

func main(){
	flag.BoolVar(&isDecode,"d",false,"Decode")
	flag.Parse()

	if isDecode {
		sc :=bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line:=sc.Text()
			buf,err:=base58.Decode(line)
			if err != nil {
				os.Stderr.WriteString(err.Error())
			} else {
				os.Stdout.Write(buf)
			}
		}
	} else {
		sc :=bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			line:=sc.Bytes()
			res:=base58.Encode(line)
			os.Stdout.WriteString(res)
		}
	}
}
