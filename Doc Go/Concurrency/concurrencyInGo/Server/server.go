package server

import (
	"fmt"
	"log"
	"net"
	"sync"

	"time"
)

var wg sync.WaitGroup

func ConnectToService() interface{} {
	time.Sleep(1 * time.Second)
	return struct{}{}
}

func warmServiceConnCache() *sync.Pool {
	p := &sync.Pool{
		New: ConnectToService,
	}
	for i := 0; i < 10; i++ {
		p.Put(p.New())
	}
	return p
}

func StartNetworkDaemon() *sync.WaitGroup {
	wg.Add(1)
	go func() {
		connPool := warmServiceConnCache()
		server, err := net.Listen("tcp", "localhost:8080")
		if err != nil {
			log.Fatalf("Cannot listen: %v", err)
		}
		defer server.Close()
		wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				log.Printf("Cannot accept connection: %v", err)
				continue
			}
			svcConn := connPool.Get()
			fmt.Println(conn, "")
			connPool.Put(svcConn)
			conn.Close()
		}
	}()
	return &wg
}
