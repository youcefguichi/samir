package main

func main() {
	c := &Container{
		Name:          "samir",
		RootFs:        "alpine-mini",
		MemoryRequest: "100Mb",
		MemoryLimit:   "500Mb",
		CpuRequest:    "100m",
		CpuLimit:      "500m",
		RunAs:         "guest",
	}

	c.Run()
}
