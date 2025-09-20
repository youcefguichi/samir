package main

import (
	"log"

	samirNet "github.com/youcef/samir/pkg/networking"
	samirRuntime "github.com/youcef/samir/pkg/runtime"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	bridge := &samirNet.BridgeSpec{
		Name:         "samir-br",
		NetworkSpace: "10.10.0.0/16",
		IP:           "10.10.0.1/16",
	}

	samirNet.CreateBridge(bridge)
	samirNet.EnableIPForwardingOnTheHost()
	samirNet.EnableNATMasquerade(bridge.Name, bridge.NetworkSpace)

	c := &samirRuntime.Container{
		Name:          "samir",
		RootFs:        "rootfs/samir-os",
		MemoryRequest: "100Mb",
		MemoryLimit:   "500Mb",
		CpuRequest:    "100m",
		CpuLimit:      "500m",
		RunAs:         "root",

		// network : "bridge",
	}

	c.Run()
}
