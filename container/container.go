package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

const (
	re_run_me         = "/proc/self/exe"
	container_command = "child_process"
	cgroup_mount_path = "/sys/fs/cgroup"
)

type Container struct {
	Name   string
	RootFs string // path to the downloaded rootfs, TODO: get the rootfs from docker hub images.

	MemoryRequest int
	MemoryLimit   int

	CpuRequest int
	CpuLimit   int

	runAs string
}

func main() {
	c := &Container{
		Name:   "samir",
		RootFs: "minialpinerootfs",
		// MemoryRequest: 10,
		// MemoryLimit:   20,
		// CpuRequest:    10,
		// CpuLimit:      20,
		runAs: "guest",
	}

	c.Run()
}

func (c *Container) Run() {

	switch os.Args[1] {
	case "run":
		c.Setup()
	case "child_process":
		c.MustSetupRootFsAndMountProcCgroup()
		c.RunContainerCommandAs("guest") // TODO: make this configurable, or run as a specific user.
	default:
		log.Fatalf("unknown command %s \n", os.Args[1])
	}

}

func (c *Container) Setup() {

	cmd := exec.Command(re_run_me, append([]string{container_command}, os.Args[2:]...)...)

	cmd.Stdin, cmd.Stdout, cmd.Stderr, cmd.SysProcAttr = os.Stdin, os.Stdout, os.Stderr, &unix.SysProcAttr{
		Cloneflags: unix.CLONE_NEWUTS | unix.CLONE_NEWPID | unix.CLONE_NEWNS,
	}

	if err := cmd.Run(); err != nil {
		log.Fatal("setting up container failed %s \n", err)
	}
}

func (c *Container) MustSetupRootFsAndMountProcCgroup() { // TODO: should i return an error or panic?

	if err := unix.Chroot(c.RootFs); err != nil {
		log.Fatal("error setting the new rootFs: %s \n", err)
	}

	if err := os.Chdir("/"); err != nil {
		log.Fatal("couldn't change the root directory to the new rootFs view %s \n", err)
	}

	// mount the proc filesystem and cgroup filesystem inside the container

	if err := unix.Mount("proc", "proc", "proc", 0, ""); err != nil {
		log.Fatal("error mounting proc fs: %s \n", err)
	}

	if err := os.MkdirAll("/sys/fs/cgroup", 0755); err != nil {
		log.Fatal("error creating /sys/fs/cgroup: %s \n", err)
	}

	err := unix.Mount("cgroup", "/sys/fs/cgroup", "cgroup2", 0, "")
	if err != nil {
		log.Printf("error mounting cgroup fs: %s \n", err)
	}

}

func (c *Container) RunContainerCommandAs(username string) {

	u, err := user.Lookup(username)
	if err != nil {
		log.Fatal("user not found: %s \n", err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		log.Fatal("error parsing UID: %s \n", err)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		log.Fatal("error parsing GID: %s \n", err)

	}

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid:         uint32(uid),
			Gid:         uint32(gid),
			NoSetGroups: true,
		},
	}

	if err := cmd.Run(); err != nil {
		log.Fatal("error running the container command: %s \n", err)
	}

	// TODO: Setup Cgroup (polish my ugly cgroup function)

}
