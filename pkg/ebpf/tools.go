package ebpf

import (
	"bufio"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"strings"
)

func checkMount(mountPath string, mountType string) (bool, error) {
	// ms, err := os.ReadFile(HostMountInfoPath)
	// if err != nil {
	// 	return false, fmt.Errorf("failed to read mount file: %v", err)
	// }
	// mss := strings.Split(string(ms), "\n")
	// for _, m := range mss {
	// 	if strings.Contains(m, fmt.Sprintf(" %s %s ", mountPath, mountType)) {
	// 		return true, nil
	// 	}
	// }
	// return false, nil

	f, err := os.Open(HostMountInfoPath)
	if err != nil {
		return false, fmt.Errorf("failed to read mount file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// example fields: cgroup2 /sys/fs/cgroup/unified cgroup2 rw,nosuid,nodev,noexec,relatime 0 0
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) >= 3 && fields[2] == mountType && strings.Compare(fields[1], mountPath) == 0 {
			return true, nil
		}
	}
	return false, nil

}

func checkOrMountCgroupV2(cgroupRoot string) error {

	if mount, err := checkMount(cgroupRoot, "cgroup2"); err != nil {
		return fmt.Errorf("failed to checkMount: %v", err)
	} else {
		if mount {
			fmt.Printf("cgroupV2 %s is already mounted \n", cgroupRoot)
			return nil
		}
	}
	fmt.Printf("begin to mount cgroupV2 fs: %s \n", cgroupRoot)

	cgroupRootStat, err := os.Stat(cgroupRoot)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cgroupRoot, 0755); err != nil {
				return fmt.Errorf("Unable to create cgroup mount directory: %w", err)
			}
		} else {
			return fmt.Errorf("Failed to stat the mount path %s: %w", cgroupRoot, err)
		}
	} else if !cgroupRootStat.IsDir() {
		return fmt.Errorf("%s is a file which is not a directory", cgroupRoot)
	}

	if err := unix.Mount("none", cgroupRoot, "cgroup2", 0, ""); err != nil {
		return fmt.Errorf("failed to mount %s: %w", cgroupRoot, err)
	}

	return nil
}

func checkOrMountBpfFs(bpfPath string) error {

	if mount, err := checkMount(bpfPath, "bpf"); err != nil {
		return fmt.Errorf("failed to checkMount: %v", err)
	} else {
		if mount {
			fmt.Printf("bpf %s is already mounted \n", bpfPath)
			return nil
		}
	}
	fmt.Printf("begin to mount bpf fs: %s \n", bpfPath)

	var err error
	_, err = os.Stat(bpfPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(bpfPath, 0755); err != nil {
				return fmt.Errorf("unable to create bpf mount directory: %s", err)
			}
		}
	}

	err = unix.Mount(bpfPath, bpfPath, "bpf", 0, "")
	if err != nil {
		return fmt.Errorf("failed to mount %s: %s", bpfPath, err)
	}

	return nil
}
