package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help")
	}
}

func run() {

	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getegid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
		Credential: &syscall.Credential{Uid: 0, Gid: 0},
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getegid())

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("./ubuntufs"))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	if _, err := os.Stat("mytemp"); os.IsNotExist(err) {
		must(os.Mkdir("mytemp", 0700))
	}

	// fmt.Println(os.Getwd())
	must(syscall.Mount("something", "mytemp", "tmpfs", 0, ""))

	must(cmd.Run())

	must(syscall.Unmount("proc", 0))
	must(syscall.Unmount("mytemp", 0))
}

func cg() {
	cgroups := "/sys/fs/cgroup"

	mem := filepath.Join(cgroups, "memory")
	fmt.Println(mem)
	os.Mkdir(filepath.Join(mem, "alexandre"), 0755)
	must(ioutil.WriteFile(filepath.Join(mem, "alexandre/memory.limit_in_bytes"), []byte("999424"), 0700))
	must(ioutil.WriteFile(filepath.Join(mem, "alexandre/memory.memsw.limit_in_bytes"), []byte("999424"), 0700))
	must(ioutil.WriteFile(filepath.Join(mem, "alexandre/memory.on_release"), []byte("1"), 0700))

	pid := strconv.Itoa(os.Getpid())
	must(ioutil.WriteFile(filepath.Join(mem, "alexandre/cgroup.procs"), []byte(pid), 0700))
}
