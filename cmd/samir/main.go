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

	samirNet.CreateBridge(bridge)
	// samirNet.EnableIPForwardingOnTheHost()
	// samirNet.EnableNATMasquerade(bridge.Name, bridge.NetworkSpace)

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
		IP:         "10.10.0.1/16",
	}

	if os.Args[1] == "run" {

		//
		samirRuntime.CreateAndConfigureCgroup(container.Resources)
		containerIface := samirNet.CreateNewVethPair(bridge.Name)
		cmd, read, write := samirRuntime.PrepareClone(namspaces)

		// pass container network specs to the child proccess
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "IP", container.IP))
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "C_IFACE", containerIface))
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "GW_IP", "10.10.0.1"))
		// cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%d", "PARENT_PID_PIPE", 3))

		if err := cmd.Start(); err != nil {
			log.Fatalf("setting up container failed %s \n", err)
		}

		read.Close()

		pid := cmd.Process.Pid

		samirNet.ApplyNetworkConfiguration(pid, containerIface, container.IP, "10.10.0.1")

		log.Printf("pid: %v", pid)

		fmt.Fprintf(write, "PID:%v", pid)
		write.Close()

		if err := cmd.Wait(); err != nil {
			log.Fatalf("container process exited with error: %v", err)
		}
	}

	if os.Args[1] == "child" {
		// samirRuntime.SetupRootFs(container.Rootfs)

		pid, err := samirRuntime.ReadDataSentByParentPID(3)

		if err != nil {
			log.Fatalf("couldn't read the data sent by the parent pid %v", err)
		}

		log.Printf("parent pid from child: %v", pid)

		// get pid
		// samirNet.ApplyNetworkConfiguration(pid)
		// err = samirRuntime.AttachInitProcessToCgroup(pid, container.Resources.Name)

		// if err != nil {
		// 	log.Fatalf("couldn't assign pid to cgroup %v", err)
		// }
		samirRuntime.SetupRootFs(container.Rootfs)
		samirRuntime.RunContainerCommandAs("root")

	}

	// samirRuntime.ManageProcessStage(os.samirNet)

}
