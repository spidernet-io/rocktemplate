package podBank

import (
	"bufio"
	"fmt"
	"github.com/spidernet-io/rocktemplate/pkg/types"
	"os"
	"path"
	"regexp"
	"strings"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// getPodAndContainerID 从给定的 cgroup 路径中提取 Pod ID 和 Container ID。
//
// 工作原理：
// 1. 打开并读取 cgroup 文件。
// 2. 使用正则表达式查找包含 "kubepods" 的行。
// 3. 解析该行以提取 Pod ID 和 Container ID。
// 4. Pod ID 通常在第四个路径段中，Container ID 在第五个路径段中。
// 5. 使用正则表达式匹配以适应不同的 cgroup 路径格式。
// 6. 将 Pod ID 中的下划线替换为连字符，以匹配 Kubernetes 中的 UID 格式。
//
// 参数：
//   - cgroupPath: cgroup 文件的路径，通常为 "/proc/<PID>/cgroup"
//
// 返回值：
//   - string: Pod ID（如果找到）
//   - string: Container ID（如果找到）
//   - bool: 是否为主机进程（如果找到）
//   - 如果未找到，两个返回值都为空字符串

func getPodAndContainerID(pid uint32) (podId string, containerId string, host bool, err error) {
	podId = ""
	containerId = ""
	host = false

	// in host
	cgroupPath := fmt.Sprintf("/proc/%d/cgroup", pid)
	if fileExists(path.Join(types.HostProcMountDir, cgroupPath)) {
		// in container
		cgroupPath = path.Join(types.HostProcMountDir, cgroupPath)
	}

	file, err := os.Open(cgroupPath)
	if err != nil {
		err = fmt.Errorf("Error opening cgroup file: %v\n", err)
		return
	}
	defer file.Close()

	podRegex := regexp.MustCompile(`kubepods-[^-]+-pod([^.]+)\.slice`)
	containerRegex := regexp.MustCompile(`[^-]+-([^.]+)\.scope`)

	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		line = scanner.Text()
		// 检查是否为主机应用
		if isHostProcess(line) {
			host = true
			return
		}
		if strings.Contains(line, "kubepods") {
			parts := strings.Split(line, "/")
			if len(parts) >= 4 {
				podMatch := podRegex.FindStringSubmatch(parts[3])
				if len(podMatch) == 2 {
					podID := strings.ReplaceAll(podMatch[1], "_", "-")

					if len(parts) >= 5 {
						containerMatch := containerRegex.FindStringSubmatch(parts[4])
						if len(containerMatch) == 2 {
							return podID, containerMatch[1], false, nil
						}
					}
				}
			}
		}
	}

	return "", "", false, fmt.Errorf("failed to get pod id of pid %d from path %s: %s", pid, cgroupPath, line)
}

func isHostProcess(line string) bool {
	hostPatterns := []string{
		"0::/",
		":/init.scope",
		":/user.slice",
		":/system.slice",
	}

	for _, pattern := range hostPatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}
