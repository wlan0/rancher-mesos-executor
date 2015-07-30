package orchestrator

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
	"golang.org/x/crypto/ssh"
)

var (
	MAX_RETRIES = 15
)

func startAndRegisterVM(imagePath, rosHDD, iface, ifaceCIDR, imageTag, imageRepo, registrationUrl, hostUuid string) error {
	ip, err := startVM(imagePath, iface, ifaceCIDR, rosHDD)
	if err != nil {
		return err
	}
	sshClient, err := getSSHClient(ip)
	if err != nil {
		return err
	}

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	session.Stdout = &stdout
	session.Stderr = &stderr

	agentCommand := buildAgentCommand(imageTag, imageRepo, registrationUrl, hostUuid)
	err = session.Run(agentCommand)

	fmt.Println(stdout.String())
	fmt.Println(stderr.String())
	if err != nil {
		return err
	}
	return nil
}

func buildAgentCommand(imageTag, imageRepo, registrationUrl, hostUuid string) string {
	return fmt.Sprintf("sudo docker run -d --privileged -v /var/run/docker.sock:/var/run/docker.sock -e CATTLE_PHYSICAL_HOST_UUID=%s %s:%s %s", hostUuid, imageRepo, imageTag, registrationUrl)
}

func startVM(image, iface, ifaceCIDR, HDD string) (net.IP, error) {
	randMac := generateRandomMacAddress()
	for i := 0; i < MAX_RETRIES; i++ {
		if ans, err := isUniqueLocalMacAddress(randMac, iface); err != nil {
			return nil, err
		} else if ans {
			break
		}
		randMac = generateRandomMacAddress()
		if i == MAX_RETRIES-1 {
			return nil, fmt.Errorf("Max retries exceeded (>%d): Too many retries to obtain a possible mac address", MAX_RETRIES)
		}
	}
	cmdStrings, err := buildKVMCommand(randMac, image, HDD)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("kvm", cmdStrings...)
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	//wait for rancherOS to boot up
	<-time.After(1 * time.Minute)

	//Hack: ping all ip addresses in subnet to fill arp-cache with the IP addresses of the newly created VM
	ip, ipnet, err := net.ParseCIDR(ifaceCIDR)
	if err != nil {
		return nil, err
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		exec.Command("ping", "-c", "1", ip.String()).Start()
		<-time.After(100 * time.Millisecond)
	}

	return waitForIp(randMac, iface)
}

func buildKVMCommand(macAddr, image, HDD string) ([]string, error) {
	var f *os.File
	var err error
	for i := 0; ; i++ {
		randStr := strconv.Itoa(i)
		if _, err = os.Stat(HDD + randStr); os.IsNotExist(err) {
			f, err = os.Create(HDD + randStr)
			if err != nil {
				return nil, fmt.Errorf("error creating file, err = %s", err)
			}
			defer f.Close()
			break
		}
	}
	hdd, err := os.Open(HDD)
	if err != nil {
		return nil, fmt.Errorf("unable to open current hdd, err = %s", err)
	}
	io.Copy(f, hdd)
	hdd.Close()
	return strings.Split(fmt.Sprintf("-drive file=%s,if=none,id=drive-disk0 -device virtio-blk-pci,scsi=off,bus=pci.0,addr=0x6,drive=drive-disk0,id=virtio-disk0,bootindex=1 -cdrom %s -boot d -netdev bridge,br=br0,id=net0 -device virtio-net-pci,netdev=net0,mac=%s -m 512m -smp 1", f.Name(), image, macAddr), " "), nil
}

func waitForIp(macAddr, iface string) (net.IP, error) {
	timeout := false
	go func() {
		<-time.After(3 * time.Minute)
		timeout = true
	}()
	for {
		ipAddr := getIPAddressFromMac(macAddr, iface)
		if ipAddr != nil {
			return ipAddr, nil
		}
		if timeout {
			return nil, fmt.Errorf("Error waiting for IP address of newly creaeted VM")
		}
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func getSSHClient(ipAddr net.IP) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: "rancher",
		Auth: []ssh.AuthMethod{
			ssh.Password("rancher"),
		},
	}
	return ssh.Dial("tcp", ipAddr.String()+":22", config)
}

func getIPAddressFromMac(macAddr, iface string) net.IP {
	br0, err := netlink.LinkByName(iface)
	if err != nil {
		return nil
	}
	neighbors, err := netlink.NeighList(br0.Attrs().Index, 0)
	if err != nil {
		return nil
	}
	for _, n := range neighbors {
		if n.HardwareAddr.String() == macAddr {
			return n.IP
		}
	}
	return nil
}

func isUniqueLocalMacAddress(macAddress, iface string) (bool, error) {
	macs, err := getMacAddresses(iface)
	if err != nil {
		return false, err
	}
	for _, addr := range macs {
		if addr == macAddress {
			return false, nil
		}
	}
	return true, nil
}

func getMacAddresses(iface string) ([]string, error) {
	br0, err := netlink.LinkByName(iface)
	if err != nil {
		return nil, fmt.Errorf("%s is not configured on this host", iface)
	}
	neighbors, err := netlink.NeighList(br0.Attrs().Index, 0)
	if err != nil {
		return nil, fmt.Errorf("neighbor list could not be retrieved")
	}
	hwAddrs := []string{}
	for _, n := range neighbors {
		hwAddrs = append(hwAddrs, n.HardwareAddr.String())
	}
	return hwAddrs, nil
}

func generateRandomMacAddress() string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return ""
	}
	// Set the local bit
	buf[0] |= 2
	return fmt.Sprintf("52:54:%02x:%02x:%02x:%02x", buf[2], buf[3], buf[4], buf[5])
}
