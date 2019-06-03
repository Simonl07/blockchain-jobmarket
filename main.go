package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"./p3"
)

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <port> <id> <firstnode_host(optional)>")
		return
	}
	nodePort := os.Args[1]
	nodeID := os.Args[2]
	var firstNodeHost string
	if len(os.Args) == 4 {
		firstNodeHost = os.Args[3]
	}
	p3.FIRST_NODE_HOST = "http://" + firstNodeHost
	p3.PORT = nodePort
	p3.NODEID = nodeID
	router := p3.NewRouter()
	fmt.Printf("Starting server on port: %v, id: %v\n", nodePort, nodeID)
	log.Fatal(http.ListenAndServe(":"+nodePort, router))
}
