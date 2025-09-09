package main

import (
	"log"
	"runtime"
	"strings"

	link "github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	samir_default_bridge_cidr = "10.10.0.0/16"
)

// host networking:
// set default bridge network to 10.10.0.0/16
// create a bridge withe the ip 10.10.0.1/16
// create veth on the bridge side
// setup a dhcp server that is binds to the bridge ip
// setup NAT

// per container networking:
// each container setup a veth
// each container request an ip from the bridg's dhcp
// add default routing to point to the bridge ip.

func CreateBridge(name string) {

	br := &link.Bridge{
		LinkAttrs: link.LinkAttrs{Name: name},
	}

	err := link.LinkAdd(br)

	if err != nil {
		log.Fatalf("couldn't create link, %v", err)
	}

	brLink, err := link.LinkByName(name)

	if err != nil {
		log.Fatalf("couldn't find link, %v", err)
	}

	addr, err := link.ParseAddr(samir_default_bridge_cidr)

	if err != nil {
		log.Fatalf("couldn't parse address %s", samir_default_bridge_cidr)
	}

	err = link.AddrAdd(brLink, addr)

	if err != nil {
		log.Fatalf("couldn't assign ip range, %s", err)
	}

	err = link.LinkSetUp(brLink)

	if err != nil {
		log.Fatalf("couldn't starts up bridge, %s", err)
	}

}

func createNewNs(ns string) {

	const (
		bridge        = "samir0"
		hostVeth      = "v0h"
		containerVeth = "v0c"
		bridgeIP      = "10.10.0.1/16"
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
