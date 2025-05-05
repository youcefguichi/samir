package main

//import "fmt"

// import (
// 	"fmt"
// 	"net"
// 	"strings"
// )

func main() {
	proxy := lb{}
	proxy.LoadConfig("config.yaml")
	proxy.start()
}
