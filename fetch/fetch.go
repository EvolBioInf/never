package main

import (
	"flag"
	"fmt"
	"github.com/evolbioinf/clio"
	"github.com/evolbioinf/never/util"
	"io"
	"net/http"
	"os"
)

func main() {
	util.PrepLog("fetch")
	u := "fetch [-v] url..."
	p := "Fetch content from one or more URLs."
	e := "fetch https://neighbors.evolbio.mpg.de/names/?t=9606"
	clio.Usage(u, p, e)
	flagV := flag.Bool("v", false, "version")
	flag.Parse()
	if *flagV {
		util.PrintInfo()
	}
	urls := flag.Args()
	if len(urls) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n",
			"plase provide at least one URL")
		os.Exit(1)
	}
	for _, url := range urls {
		res, err := http.Get(url)
		util.Check(err)
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		util.Check(err)
		fmt.Printf("%s", body)
	}
}
