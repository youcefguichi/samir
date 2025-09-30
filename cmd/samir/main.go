package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	samirNet "github.com/youcef/samir/pkg/networking"
	samirRuntime "github.com/youcef/samir/pkg/runtime"
	"golang.org/x/sys/unix"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

const (
	namspaces = unix.CLONE_NEWUTS | unix.CLONE_NEWPID | unix.CLONE_NEWNS | unix.CLONE_NEWNET

	// networking defaults
	br_name          = "samir-br"
	br_ip            = "10.10.0.1/16"
	br_network_space = "10.10.0.0/16"
)

func main() {

	bridge := &samirNet.BridgeSpec{
		Name:         br_name,
		IP:           br_ip,
		NetworkSpace: br_network_space,
	}

	if os.Args[1] != "child" {
		samirNet.CreateBridge(bridge)
		samirNet.EnableIPForwardingOnTheHost()
		samirNet.EnableNATMasquerade(bridge.Name, bridge.NetworkSpace)
	}

	data, err := os.ReadFile("bundle/config.json")

	var container samirRuntime.ContainerSpec
	if err != nil {
		log.Printf("open file %v", err)
	}

	err = json.Unmarshal(data, &container)
	if err != nil {
		log.Fatalf("Unmarshel json %v", err)
	}

	if os.Args[1] == "run" {

		r, w, err := os.Pipe()

		if err != nil {
			log.Fatalf("create pipe %v", err)
		}

		samirRuntime.CreateAndConfigureCgroup(container.Cgroup)
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
		err = samirRuntime.AttachInitProcessToCgroup(pid, container.Cgroup.Name)

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
