package main

import (
	"fmt"
	"log"
	"os"

	// "time"

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
		MaxMem: "10Mb",
		MinMem: "10Mb",
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

		r, w, err := os.Pipe()

		if err != nil {
			log.Fatalf("create pipe %v", err)
		}

		samirRuntime.CreateAndConfigureCgroup(container.Resources)
		containerIface := samirNet.CreateNewVethPair(bridge.Name)
		cmd := samirRuntime.PrepareClone(namspaces)

		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "IP", container.IP))
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "C_IFACE", containerIface))
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "GW_IP", "10.10.0.1"))
		cmd.ExtraFiles = []*os.File{r} // pass the read end of the pipe to child process

		if err := cmd.Start(); err != nil {
			log.Fatalf("setting up container failed %s \n", err)
		}

		pid := cmd.Process.Pid
		samirNet.MoveVethToNetworkNamespace(pid, containerIface)
		err = samirRuntime.AttachInitProcessToCgroup(pid, container.Resources.Name)

		if err != nil {
			log.Fatalf("couldn't assign pid to cgroup %v", err)
		}

		r.Close()
		fmt.Fprintf(w, "%d", 1) // signal to the child process to start to avoid race condition.
		w.Close()

		if err := cmd.Wait(); err != nil {
			log.Fatalf("container process exited with error: %v", err)
		}
	}

	if os.Args[1] == "child" {

		err := samirRuntime.WaitForSignalFromParentPID(3)
		if err != nil {
			log.Fatalf("receiving signal from parent %v \n", err)
		}

		samirNet.MustSetupContainerInterface()
		samirRuntime.SetupRootFs(container.Rootfs)
		samirRuntime.RunContainerCommandAs("root")
	}

}
