package main

import (
	"log"
	"runtime"
	"strings"

	link "github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	samir_bridge_default_name = "samir0"
	bridge_ip                 = "10.10.0.1/16"
	sh0                       = "sh0"
	sc0                       = "sco"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

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

	addr, err := link.ParseAddr(bridge_ip)

	if err != nil {
		log.Fatalf("couldn't parse address %s", bridge_ip)
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

func SetupVeth(br string, sho string, sco string) {

	veth := &link.Veth{
		LinkAttrs: link.LinkAttrs{Name: sho},
		PeerName:  sco,
	}

	err := link.LinkAdd(veth)

	if err != nil {
		log.Fatalf("create veth pair: %v", err)
	}

	hostVeth, err := link.LinkByName(sho)

	if err != nil {
		log.Fatalf("lookup host veth: %v", err)
	}

	bridge, err := link.LinkByName(br)

	if err != nil {
		log.Fatalf("lookup bridge: %v", err)
	}

	err = link.LinkSetMaster(hostVeth, bridge)

	if err != nil {
		log.Fatalf("set host veth master %v", err)
	}

	err = link.LinkSetUp(hostVeth)

	if err != nil {
		log.Fatalf("bring host veth up %s", err)
	}

}

func ConfigureHostNetworking() {
	CreateBridge(samir_bridge_default_name)
	SetupVeth(samir_bridge_default_name, sh0, sc0)
}

func CreateNewNs(ns string) {

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
