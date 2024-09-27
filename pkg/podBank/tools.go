package podBank

import (
	"fmt"
	"github.com/spidernet-io/rocktemplate/pkg/types"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"path"
)

type IPAddresses struct {
	IPv4 []net.IP
	IPv6 []net.IP
}

func GetContainerIP(pid int, interfaceNames []string) (*IPAddresses, error) {
	// Save current network namespace
	currentNS, err := netns.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get current network namespace: %v", err)
	}
	defer currentNS.Close()

	// Get target process network namespace
	targetNS, err := netns.GetFromPid(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to get target process network namespace: %v", err)
	}
	defer targetNS.Close()

	var allIPs IPAddresses

	// Switch to target network namespace
	err = netns.Set(targetNS)
	if err != nil {
		return nil, fmt.Errorf("failed to switch to target network namespace: %v", err)
	}

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	for _, iface := range interfaces {
		// Skip loopback interface
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// If interface names are specified, only process those
		if len(interfaceNames) > 0 && !containStr(interfaceNames, iface.Name) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("failed to get the ip of interface %s: %v", iface.Name, err)
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP

			// Filter out link-local addresses
			if ip.IsLinkLocalUnicast() {
				continue
			}

			if ip.To4() != nil {
				if !containsIP(allIPs.IPv4, ip) {
					allIPs.IPv4 = append(allIPs.IPv4, ip)
				}
			} else {
				if !containsIP(allIPs.IPv6, ip) {
					allIPs.IPv6 = append(allIPs.IPv6, ip)
				}
			}
		}
	}

	// Switch back to original network namespace
	err = netns.Set(currentNS)
	if err != nil {
		return nil, fmt.Errorf("failed to switch back to original network namespace: %v", err)
	}

	if len(allIPs.IPv4) == 0 && len(allIPs.IPv6) == 0 {
		return nil, fmt.Errorf("no valid IP addresses found")
	}

	return &allIPs, nil
}

func containStr(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsIP(slice []net.IP, ip net.IP) bool {
	for _, a := range slice {
		if a.Equal(ip) {
			return true
		}
	}
	return false
}

func CheckHostNetNamespaceByPid(pid int) (bool, error) {

	// detect that it is running in a container or on the host
	nsPath := path.Join(types.HostProcMountDir, "/proc/1/ns/net")
	if _, err := os.Stat(nsPath); os.IsNotExist(err) {
		nsPath = "/proc/1/ns/net"
	}
	hostNS, err := netns.GetFromPath(nsPath)
	if err != nil {
		return false, fmt.Errorf("failed to get host network namespace: %v", err)
	}
	defer hostNS.Close()

	targetNS, err := netns.GetFromPid(pid)
	if err != nil {
		return false, fmt.Errorf("failed to get target process network namespace: %v", err)
	}
	defer targetNS.Close()

	return hostNS.Equal(targetNS), nil
}
