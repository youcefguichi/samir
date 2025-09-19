package runtime

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	re "regexp"
	"strconv"
	"strings"
	"syscall"

	link "github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

const (
	re_run_me                 = "/proc/self/exe"
	container_command         = "child_process"
	cgroup_mount_path         = "/sys/fs/cgroup"
	proc_mount_info           = "/proc/self/mountinfo"
	samir_bridge_default_name = "samir0"
	bridge_ip                 = "10.10.0.1/16"
	sh0                       = "sh0"
	sc0                       = "sc0"
)

type Container struct {
	Name   string
	RootFs string // path to the downloaded rootfs, TODO: get the rootfs from docker hub images.

	MemoryRequest string // e.g. "10Mb"
	MemoryLimit   string

	CpuRequest string // e.g. "10m"
	CpuLimit   string

	RunAs string
}

func (c *Container) Run() {

	switch os.Args[1] {
	case "run":
		c.Init()
	case "child_process":
		c.SetupRootFsWithProcAndCgroupMounts()
		c.RunContainerCommandAs(c.RunAs) // TODO: make this configurable, or run as a specific user.
	default:
		log.Fatalf("unknown command %s \n", os.Args[1])
	}

}

func (c *Container) Init() {
	err := createNewCgroup(c.Name)

	if os.IsNotExist(err); err != nil {
		log.Printf("cgroup  with name '%s' already exist skipping creation .. \n", c.Name)
	} else {
		log.Printf("couldn't create cgroup '%s' due to %s \n", c.Name, err)
	}

	veth := &link.Veth{
		LinkAttrs: link.LinkAttrs{Name: "sh-4"}, // TODO: generate automatically
		PeerName:  "sc-4",
	}

	MustCreateVethPair("samir-br", veth)

	c.CloneAndConfigureContainerNetworking(unix.CLONE_NEWUTS|unix.CLONE_NEWPID|unix.CLONE_NEWNS|unix.CLONE_NEWNET, veth, veth.PeerName)
}

func (c *Container) CloneAndConfigureContainerNetworking(namespaces uintptr, veth *link.Veth, sc0 string) {

	cmd := exec.Command(re_run_me, append([]string{container_command}, os.Args[2:]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr, cmd.SysProcAttr = os.Stdin, os.Stdout, os.Stderr, &unix.SysProcAttr{
		Cloneflags: namespaces,
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("setting up container failed %s \n", err)
	}

	hostPID := cmd.Process.Pid
	log.Printf("Container init host PID: %d", hostPID)

	// EYY WORKS!!
	cns := &ContainerNetworkSpec{
		ContainerPID:  hostPID,
		ContainerVeth: sc0,
		IP:            "10.10.0.12/16",
		GatewayIP:     "10.10.0.1",
	}

	MustSetupContainerNetwork(cns)

	if err := cmd.Wait(); err != nil {
		log.Fatalf("container process exited with error: %v", err)
	}
}

func (c *Container) SetupRootFsWithProcAndCgroupMounts() { // TODO: should i return an error or panic?
	fmt.Printf("Container setup rootfs PID: %d\n", os.Getpid())
	if cwd, err := os.Getwd(); err != nil {
		log.Printf("warning: getwd failed: %v", err)
	} else {
		log.Printf("current working directory before chroot: %s", cwd)
	}

	if err := unix.Chroot(c.RootFs); err != nil {
		log.Fatalf("error setting the new rootFs: %s \n", err)
	}

	if err := os.Chdir("/"); err != nil {
		log.Fatalf("couldn't change the root directory to the new rootFs view %s \n", err)
	}

	if err := unix.Mount("proc", "proc", "proc", 0, ""); err != nil {
		log.Fatalf("error mounting proc fs: %s \n", err)
	}

	_, err := os.Stat(cgroup_mount_path)
	if os.IsNotExist(err) {
		log.Printf("cgroup directory does not exist, creating it...\n")
		if err := os.MkdirAll(cgroup_mount_path, 0777); err != nil {
			log.Fatalf("error creating %s: %s \n", cgroup_mount_path, err)
		}
	}

	if !isMounted(cgroup_mount_path) {
		err = unix.Mount("cgroup", cgroup_mount_path, "cgroup2", 0, "")
		if err != nil {
			log.Printf("error mounting cgroup fs: %s \n", err.Error())
		}
	}

	// Set resource limits and requests

	if err := setMemoryLimits(c.MemoryLimit, c.Name); err != nil {
		log.Fatalf("error setting memory limit for cgroup %s: %s \n", c.Name, err)
	}

	if err := setMemoryRequests(c.MemoryRequest, c.Name); err != nil {
		log.Fatalf("error setting memory request for cgroup %s: %s \n", c.Name, err)
	}

	if err := setCpuLimits(c.CpuLimit, c.Name); err != nil {
		log.Fatalf("error setting cpu limit for cgroup %s: %s \n", c.Name, err)
	}

	if err := setCpuRequests(c.CpuRequest, c.Name); err != nil {
		log.Fatalf("error setting cpu request for cgroup %s: %s \n", c.Name, err)
	}

	assignPidToCgroup(os.Getpid(), c.Name)

}

func (c *Container) RunContainerCommandAs(username string) {

	if !isRootOrGuest(username) {
		log.Fatalf("you can only run the container command as 'root' or 'guest', got '%s' \n", username)
	}

	u, err := user.Lookup(username)
	if err != nil {
		log.Fatalf("user not found: %s \n", err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		log.Fatalf("error parsing UID: %s \n", err)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		log.Fatalf("error parsing GID: %s \n", err)

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
		log.Fatalf("error running the container command: %s \n", err)
	}

}

func createNewCgroup(name string) error {
	cgroup_path := cgroup_mount_path + "/" + name
	return os.Mkdir(cgroup_path, 0755)
}

func assignPidToCgroup(pid int, cgroupName string) error {
	cgroupPath := cgroup_mount_path + "/" + cgroupName
	return os.WriteFile(cgroupPath+"/cgroup.procs", []byte(strconv.Itoa(pid)), 0644)
}

func setMemoryLimits(value string, cgroupName string) error {
	v, err := ValidateMemoryInputAndExtractValue(value)
	if err != nil {
		return fmt.Errorf("error validating memory value: %s, %w", value, err)
	}

	memoryLimitPath := cgroup_mount_path + "/" + cgroupName + "/memory.max"

	log.Printf("setting memory limit to %sMb for cgroup %s\n", v, cgroupName)

	v = convertFromMbtoBytes(v)
	return os.WriteFile(memoryLimitPath, []byte(v), 0644)
}

func setMemoryRequests(value string, cgroupName string) error {
	v, err := ValidateMemoryInputAndExtractValue(value)
	if err != nil {
		return fmt.Errorf("error validating memory request value: %s, %w", value, err)
	}
	memoryRequestPath := cgroup_mount_path + "/" + cgroupName + "/memory.low"

	log.Printf("setting memory request to %sMb for cgroup %s\n", v, cgroupName)

	v = convertFromMbtoBytes(v)
	return os.WriteFile(memoryRequestPath, []byte(v), 0644)
}

func setCpuLimits(value string, cgroupName string) error {
	v, err := validateCpuInputAndExtractValue(value)
	if err != nil {
		return fmt.Errorf("error validating cpu limit value: %s, %w", value, err)
	}

	cpuLimitPath := cgroup_mount_path + "/" + cgroupName + "/cpu.max"

	log.Printf("setting cpu limit to %sm for cgroup %s\n", v, cgroupName)

	v = convertMillicoresToCpuSpec(v)
	return os.WriteFile(cpuLimitPath, []byte(v), 0644)
}

func setCpuRequests(value string, cgroupName string) error {
	v, err := validateCpuInputAndExtractValue(value)
	if err != nil {
		return fmt.Errorf("error validating cpu request value: %s, %w", value, err)
	}

	cpuRequestPath := cgroup_mount_path + "/" + cgroupName + "/cpu.weight"

	log.Printf("setting cpu request to %sm for cgroup %s\n", v, cgroupName)

	//v = convertMillicoresToCpuSpec(v)
	return os.WriteFile(cpuRequestPath, []byte(v), 0644)
}

func isMounted(target string) bool {

	file, err := os.Open(proc_mount_info)
	if err != nil {
		return false
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		if fields[4] == target {
			return true
		}
	}
	return false
}

func isRootOrGuest(username string) bool {
	return username == "root" || username == "guest"
}

func validateCpuInputAndExtractValue(cpuValue string) (string, error) {

	re := re.MustCompile(`^(\d+)m$`)
	res := re.FindStringSubmatch(cpuValue)
	if res == nil {
		return "", fmt.Errorf("invalid cpu value format: %s, expected format is <number>m", cpuValue)
	}

	return res[1], nil
}

func ValidateMemoryInputAndExtractValue(memoryValue string) (string, error) {

	re := re.MustCompile(`^(\d+)Mb$`)
	res := re.FindStringSubmatch(memoryValue)
	if res == nil {
		return "", fmt.Errorf("invalid memory value format: %s, expected format is <number>Mb", memoryValue)
	}

	return res[1], nil
}

func convertFromMbtoBytes(memoryMB string) string {
	v, err := strconv.ParseInt(memoryMB, 10, 64)

	if err != nil {
		panic(err)
	}

	if v < 0 {
		panic("memory value cannot be negative")
	}

	valueInBytes := v * 1024 * 1024
	return strconv.FormatInt(valueInBytes, 10)
}

func convertMillicoresToCpuSpec(millicores string) string {
	v, err := strconv.ParseInt(millicores, 10, 64)

	if err != nil {
		panic(err)
	}

	if v < 0 {
		panic("cpu value cannot be negative")
	}

	period := int64(100000) // 100ms in microseconds : this is the default period for cgroup v2
	quota := (v * period) / 1000

	return fmt.Sprintf("%d %d", quota, period)
}
