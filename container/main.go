package main

import "log"

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	bridge := &BridgeSpec{
		Name:         "samir-br",
		NetworkSpace: "10.10.0.0/16",
		IP:           "10.10.0.1/16",
	}

	CreateBridge(bridge)

	c := &Container{
		Name:          "samir",
		RootFs:        "alpine-mini-v3",
		MemoryRequest: "100Mb",
		MemoryLimit:   "500Mb",
		CpuRequest:    "100m",
		CpuLimit:      "500m",
		RunAs:         "root",

		// network : "bridge",
	}

	c.Run()
}
