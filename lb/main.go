package main

//import "fmt"

// import (
// 	"fmt"
// 	"net"
// 	"strings"
// )

func main() {
	lb := lb{}
	lb.LoadConfig("config.yaml")
	lb.start()
}
