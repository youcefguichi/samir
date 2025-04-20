package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {
	lb := lb{}
	lb.LoadConfig("config.yaml")
	lb.start()
}


func parseCIDR(cidr string) {
   _, netIP, err  := net.ParseCIDR(cidr)

   if err != nil {
	  fmt.Println("Error parsing CIDR:", err)
	  return
   }

   fmt.Println(netIP.Contains(net.ParseIP("10.0.0.1")))

   strings.HasPrefix()

}
