package networking

import (
	"fmt"
	"log"
	"net"
	"os/exec"

	uid "github.com/google/uuid"
	link "github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
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

type BridgeSpec struct {
	ID           int
	Name         string
	NetworkSpace string
	// BridgeInterface    string
	// ContainerInterface string
	IP string
	// HostVeth           *link.Veth
}

// type ContainerNetworkSpec struct {
// 	IP            string
// 	ContainerPID  int
// 	ContainerVeth string
// 	Interface     string
// 	GatewayIP     string
// 	DHCPServer    string
// 	DefaultRoute  *link.Route
// }

func CreateBridge(bridge *BridgeSpec) {

	br := &link.Bridge{
		LinkAttrs: link.LinkAttrs{Name: bridge.Name},
	}

	err := link.LinkAdd(br)

	if err != nil {
		log.Printf("couldn't create link, %v", err)
	}

	brLink, err := link.LinkByName(bridge.Name)

	if err != nil {
		log.Fatalf("couldn't find link, %v", err)
	}

	addr, err := link.ParseAddr(bridge.IP)

	if err != nil {
		log.Fatalf("couldn't parse address %s", bridge.IP)
	}

	err = link.AddrAdd(brLink, addr)

	if err != nil {
		log.Printf("couldn't assign ip range, %s", err)
	}

	err = link.LinkSetUp(brLink)

	if err != nil {
		log.Fatalf("couldn't starts up bridge, %s", err)
	}

}

func ApplyNetworkConfiguration(pid int, c_iface string, ip string, gw_ip string) {
    
    // this function should be splitted


	// // get network specs
	// ip := os.Getenv("IP")
	// c_iface := os.Getenv("C_IFACE")
	// gw_ip := os.Getenv("GW_IP")
	log.Printf("IP: %s, C_IFACE: %s, GW_IP: %s", ip, c_iface, gw_ip)

	nsHandle, err := netns.GetFromPid(pid)

	if err != nil {
		log.Printf("couldn't get namespace %v", err)
	}

	defer nsHandle.Close()

	vethNetworkLink, err := link.LinkByName(c_iface)

	if err != nil {
		log.Printf("couldn't get link %v", err)
	}

	err = link.LinkSetNsFd(vethNetworkLink, int(nsHandle))

	if err != nil {
		log.Printf("could not set netns for container veth: %v", err)
	}

	MustSetupContainerInterface(
		pid,
		c_iface,
		ip,
		gw_ip,
	)
}

func CreateNewVethPair(br string) string {

	veth := PrepareNewVethObject()

	err := link.LinkAdd(veth)

	if err != nil {
		log.Printf("create veth pair: %v", err)
	}

	hostVeth, err := link.LinkByName(veth.Name)

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

	return veth.PeerName
}

func PrepareNewVethObject() *link.Veth {
	id := uid.New().String()[:4] // TODO: what is the best way to do this
	veth := &link.Veth{
		LinkAttrs: link.LinkAttrs{Name: "sh-" + id},
		PeerName:  "sc-" + id,
	}

	return veth
}
func MoveVethToNetworkNamespace(ifaceName string, pid int) error {
	// Find the link by name in the host namespace
	l, err := link.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to find link %s in host namespace: %w", ifaceName, err)
	}

	// Move the link into the child's network namespace
	if err := link.LinkSetNsPid(l, pid); err != nil {
		return fmt.Errorf("failed to set link netns for pid %d: %w", pid, err)
	}
	return nil
}

func MustSetupContainerInterface(pid int, ifName string, IP string, GwIP string) {

	nsHandle, err := netns.GetFromPid(pid)

	if err != nil {
		log.Fatalf("could not get netns for pid %d: %v", pid, err)
	}

	defer nsHandle.Close()

	currentNS, err := netns.Get()
	if err != nil {
		log.Fatalf("could not get current netns: %v", err)
	}
	defer currentNS.Close()

	if err := netns.Set(nsHandle); err != nil {
		log.Fatalf("could not set netns: %v", err)
	}
	defer netns.Set(currentNS)

	Netnslink, err := link.LinkByName(ifName)
	if err != nil {
		log.Fatalf("could not get link %s: %v", ifName, err)
	}

	addr, err := link.ParseAddr(IP) // TODO: request the IP FROM A DHCP SERVER
	if err != nil {
		log.Fatalf("could not parse IP address: %v", err)
	}

	if err := link.AddrAdd(Netnslink, addr); err != nil {
		log.Fatalf("could not add IP address: %v", err)
	}

	if err := link.LinkSetUp(Netnslink); err != nil {
		log.Fatalf("could not bring up interface: %v", err)
	}

	// Add default route to the gateway
	gw := net.ParseIP(GwIP)
	route := &link.Route{
		LinkIndex: Netnslink.Attrs().Index,
		Gw:        gw,
	}

	if err := link.RouteAdd(route); err != nil {
		log.Fatalf("could not add route: %v", err)
	}

}

func EnableIPForwardingOnTheHost() {

	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	err := cmd.Run()

	if err != nil {
		log.Fatalf("enable ip_forward: %v ", err)
	}
}

func EnableNATMasquerade(bridge string, networkSpace string) {

	// for now the full nat config is under scripts/nat.sh
	// TODO: move it here
	cmd := exec.Command(
		"iptables",
		"-t",
		"nat",
		"-A",
		"POSTROUTING",
		"-s",
		networkSpace,
		"!",
		"-o",
		bridge,
		"-j",
		"MASQUERADE")

	err := cmd.Run()

	if err != nil {
		log.Fatalf("NAT Masquerade: %v ", err)
	}
}
