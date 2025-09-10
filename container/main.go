package main

func main() {
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
