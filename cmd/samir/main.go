package main

import (
	"fmt"
	"log"
	"os"

	uid "github.com/google/uuid"
	samirNet "github.com/youcef/samir/pkg/networking"
	samirRuntime "github.com/youcef/samir/pkg/runtime"
	"golang.org/x/sys/unix"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

const (
	namspaces = unix.CLONE_NEWUTS | unix.CLONE_NEWPID | unix.CLONE_NEWNS | unix.CLONE_NEWNET
)

func main() {

	bridge := &samirNet.BridgeSpec{
		Name:         "samir-br",
		NetworkSpace: "10.10.0.0/16",
		IP:           "10.10.0.1/16",
	}

	if os.Args[1] != "child" {
		samirNet.CreateBridge(bridge)
		samirNet.EnableIPForwardingOnTheHost()
		samirNet.EnableNATMasquerade(bridge.Name, bridge.NetworkSpace)
	}

	// TODO: the limits are not being enforced check the race condition with the parent child
	resources := &samirRuntime.CgroupSpec{
		Name:   "samir",
		MaxMem: "1Mb",
		MinMem: "1Mb",
		MaxCPU: "100m",
		MinCPU: "500m",
	}

	id := uid.New()

	container := &samirRuntime.ContainerSpec{
		ID:         id.String()[:8],
		Name:       "samir",
		Rootfs:     "rootfs/samir-os",
		Entrypoint: []string{"/bin/sh"},
		Resources:  resources,
		RunAs:      "root",
		IP:         "10.10.0.4/16",
	}

	if os.Args[1] == "run" {

		samirRuntime.CreateAndConfigureCgroup(container.Resources)
		containerIface := samirNet.CreateNewVethPair(bridge.Name)
		cmd, _, _ := samirRuntime.PrepareClone(namspaces)

		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "IP", container.IP))
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "C_IFACE", containerIface))
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "GW_IP", "10.10.0.1"))

		if err := cmd.Start(); err != nil {
			log.Fatalf("setting up container failed %s \n", err)
		}

		pid := cmd.Process.Pid
		samirNet.MoveVethToNetworkNamespace(pid, containerIface)
		err := samirRuntime.AttachInitProcessToCgroup(pid, container.Resources.Name)

		if err != nil {
			log.Fatalf("couldn't assign pid to cgroup %v", err)
		}

		if err := cmd.Wait(); err != nil {
			log.Fatalf("container process exited with error: %v", err)
		}
	}

	if os.Args[1] == "child" {
		samirNet.MustSetupContainerInterface()
		samirRuntime.SetupRootFs(container.Rootfs)
		samirRuntime.RunContainerCommandAs("root")
	}

}
