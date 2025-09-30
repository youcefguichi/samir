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

	"golang.org/x/sys/unix"
)

const (
	re_run_me                 = "/proc/self/exe"
	cgroup_mount_path         = "/sys/fs/cgroup"
	proc_mount_info           = "/proc/self/mountinfo"
	samir_bridge_default_name = "samir0"
	bridge_ip                 = "10.10.0.1/16"
	sh0                       = "sh0"
	sc0                       = "sc0"
)

// build container
// build the image using docker
// samir would extract to image to a specific location and use it as a rootfs

// runtime (where saming is doing its job)
// the first run:
// create and setup cgroup
// clone the namespaces
// the second run:
// perform chroot
// change the working directory
// mount /proc filesystem

type ContainerSpec struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Rootfs     string      `json:"rootfs"`
	Entrypoint []string    `json:"entrypoint"`
	CMD        []string    `json:"cmd"`
	RunAs      string      `json:"run_as"`
	Cgroup     *CgroupSpec `json:"cgroup"`
	IP         string      `json:"ip"`
}

type CgroupSpec struct {
	Name   string `json:"name"`
	MaxMem string `json:"max_mem"`
	MinMem string `json:"min_mem"`
	MaxCPU string `json:"max_cpu"`
	MinCPU string `json:"min_cpu"`
}

func PrepareClone(namespaces uintptr) *exec.Cmd {

	cmd := exec.Command(re_run_me, append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr, cmd.SysProcAttr = os.Stdin, os.Stdout, os.Stderr, &unix.SysProcAttr{
		Cloneflags: namespaces,
	}

	return cmd
}

func SetupRootFs(rootfs string) { // TODO: should i return an error or panic?

	if err := unix.Chroot(rootfs); err != nil {
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

}

func RunContainerCommandAs(username string) {

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

func CreateAndConfigureCgroup(c *CgroupSpec) {

	cgroup_path := cgroup_mount_path + "/" + c.Name
	err := os.Mkdir(cgroup_path, 0755)

	if err != nil {
		log.Printf("couldn't create cgroup %v", err)
	}

	if err := SetMaxMem(c.MaxMem, c.Name); err != nil {
		log.Fatalf("error setting memory limit for cgroup %s: %s \n", c.Name, err)
	}

	if err := SetMinMem(c.MinMem, c.Name); err != nil {
		log.Fatalf("error setting memory request for cgroup %s: %s \n", c.Name, err)
	}

	if err := setMaxCPU(c.MaxCPU, c.Name); err != nil {
		log.Fatalf("error setting cpu limit for cgroup %s: %s \n", c.Name, err)
	}

	if err := SetMinCPU(c.MinCPU, c.Name); err != nil {
		log.Fatalf("error setting cpu request for cgroup %s: %s \n", c.Name, err)
	}

}

func AttachInitProcessToCgroup(pid int, cgroupName string) error {
	cgroupPath := cgroup_mount_path + "/" + cgroupName
	return os.WriteFile(cgroupPath+"/cgroup.procs", []byte(strconv.Itoa(pid)), 0644)
}

func SetMaxMem(value string, cgroupName string) error {
	v, err := ValidateMemoryInputAndExtractValue(value)
	if err != nil {
		return fmt.Errorf("error validating memory value: %s, %w", value, err)
	}

	memoryLimitPath := cgroup_mount_path + "/" + cgroupName + "/memory.max"

	log.Printf("setting memory limit to %sMb for cgroup %s\n", v, cgroupName)

	v = convertFromMbtoBytes(v)
	return os.WriteFile(memoryLimitPath, []byte(v), 0644)
}

func SetMinMem(value string, cgroupName string) error {
	v, err := ValidateMemoryInputAndExtractValue(value)
	if err != nil {
		return fmt.Errorf("error validating memory request value: %s, %w", value, err)
	}
	memoryRequestPath := cgroup_mount_path + "/" + cgroupName + "/memory.low"

	log.Printf("setting memory request to %sMb for cgroup %s\n", v, cgroupName)

	v = convertFromMbtoBytes(v)
	return os.WriteFile(memoryRequestPath, []byte(v), 0644)
}

func setMaxCPU(value string, cgroupName string) error {
	v, err := validateCpuInputAndExtractValue(value)
	if err != nil {
		return fmt.Errorf("error validating cpu limit value: %s, %w", value, err)
	}

	cpuLimitPath := cgroup_mount_path + "/" + cgroupName + "/cpu.max"

	log.Printf("setting cpu limit to %sm for cgroup %s\n", v, cgroupName)

	v = convertMillicoresToCpuSpec(v)
	return os.WriteFile(cpuLimitPath, []byte(v), 0644)
}

func SetMinCPU(value string, cgroupName string) error {
	v, err := validateCpuInputAndExtractValue(value)
	if err != nil {
		return fmt.Errorf("error validating cpu request value: %s, %w", value, err)
	}

	cpuRequestPath := cgroup_mount_path + "/" + cgroupName + "/cpu.weight"

	log.Printf("setting cpu request to %sm for cgroup %s\n", v, cgroupName)

	//v = convertMillicoresToCpuSpec(v)
	return os.WriteFile(cpuRequestPath, []byte(v), 0644)
}

func WaitForSignalFromParentPID(fd int) error {

	pipe := os.NewFile(uintptr(3), "parent_pid_pipe")
	defer pipe.Close()

	reader := bufio.NewReader(pipe)
	line, err := reader.ReadString(1)

	if err != nil && !strings.Contains(err.Error(), "EOF") {
		return err
	}

	log.Printf("received start signal from the parent pid [SIGNAL=%v] \n", line)

	return nil
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
