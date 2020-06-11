package main

import (
	"fmt"
	"github.com/mr-tron/base58"
	"github.com/yottachain/YTDataNode/config"
	"github.com/yottachain/YTDataNode/util"
	"github.com/yottachain/YTFS/storage"
	"log"
	"os"
	"path"
)

var ytpath = util.GetYTFSPath()

func main() {


	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	yt, err := storage.GetTableIterator(path.Join(ytpath, "index.db"), cfg.Options)
	if err != nil {
		log.Fatal(err)
	}
	var i uint64 = 0
	for {
		i++
		tb, err := yt.GetNoNilTable()
		if err != nil {
			return
		}
		for k, _ := range tb {
			fmt.Fprintln(os.Stdout, base58.Encode(k[:]))
		}
	}
}
