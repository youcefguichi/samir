package main

import (
	"log"
	"runtime"
	"strings"

	link "github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	ns "github.com/vishvananda/netns"
)

func createNewNs(ns string) {

	const (
		bridge        = "samir0"
		hostVeth      = "v0h"
		containerVeth = "v0c"
		bridgeIP      = "10.10.0.1/24"
	)

	// when settin up operations like namespaces
	// locking the thread the current running goroutine is a must
	// since the gouroutine can be scheduled into a different os thread
	// this may lead to creating a network operations on a different namespace
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	newNs, err := netns.NewNamed(ns)

	if err != nil {

		if strings.Contains(err.Error(), "file exists") {
			newNs, err = netns.GetFromName(ns)
			if err != nil {
				log.Fatalf("could no open existing ns %s, %v", ns, err)
			}
		}

		log.Fatalf("failed to create netns %s: %v", ns, err)
	}

	defer newNs.Close()

}
