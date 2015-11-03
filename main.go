package main

import (
	"flag"
	"github.com/mconintet/kiwi"
	"log"
	"net"
)

func main() {
	flag.StringVar(&dbConnInfo, "db", "", "database access info")
	var srvAddr string
	flag.StringVar(&srvAddr, "sa", ":9876", "server address")

	flag.Parse()
	if dbConnInfo == "" {
		flag.Usage()
		log.Fatal("invalid config: missing databse access info")
	}

	srv := kiwi.NewServer()
	srv.Addr, _ = net.ResolveTCPAddr("tcp", srvAddr)

	srv.ApplyDefaultCfg()

	srv.OnConnOpenFunc("/", onConnOpen)
	srv.OnConnCloseFunc("/", onConnClose)

	log.Println("Server is running on [" + srvAddr + "]")
	srv.ListenAndServe()
}
