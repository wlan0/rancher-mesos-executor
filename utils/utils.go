package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
)

var (
	rancherOSURL = "http://github.com/rancher/os/releases/download/v0.3.3/rancheros.iso"
	baseImageURL = "https://s3-us-west-1.amazonaws.com/wlan0/base-img.img"
)

func PerformPreChecksAndPrepareHost(workDir string) error {
	if !InSupportedOS() {
		return fmt.Errorf("Unsupported OS : %s, should be one among %v", CurrentOS(), SupportedOSes())
	}
	_, err := exec.LookPath("virsh")
	if err != nil {
		log.Info("KVM not found in PATH. Installing KVM")
		err := InstallKVM(CurrentOS())
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(filepath.Join(workDir, "rancheros.iso")); os.IsNotExist(err) {
		return Download(workDir, "rancheros.iso", rancherOSURL)
	}
	if _, err := os.Stat(filepath.Join(workDir, "base-img.img")); os.IsNotExist(err) {
		return Download(workDir, "base-img.img", baseImageURL)
	}
	return nil
}

func Download(dir, file, isoUrl string) error {
	client := http.Client{}
	s, err := client.Get(isoUrl)
	if err != nil {
		return err
	}
	src := s.Body
	defer src.Close()

	f, err := ioutil.TempFile(dir, file+".tmp")
	if err != nil {
		return err
	}
	defer func() {
		if _, err := os.Stat(f.Name()); !os.IsNotExist(err) {
			os.Remove(f.Name())
		}
	}()
	_, err = io.Copy(f, src)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	dest := filepath.Join(dir, file)
	return os.Rename(f.Name(), dest)

}

func InstallKVM(currentOS string) error {
	var cmd *exec.Cmd
	switch currentOS {
	case "ubuntu":
		cmd = exec.Command("apt-get", "install", "-y", "qemu-kvm", "libvirt-bin", "ubuntu-vm-builder", "bridge-utils")
	default:
		fmt.Errorf("Unsupported OS")
	}
	return cmd.Run()
}

func CurrentOS() string {
	osReleaseOut, err := exec.Command("cat", "/etc/os-release").Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(osReleaseOut), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID=") {
			if len(line) > 3 {
				return line[3:]
			}
		}
	}
	return ""
}

func InSupportedOS() bool {
	return supports(CurrentOS())
}

func SupportedOSes() []string {
	return []string{"ubuntu"}
}

func supports(currentOS string) bool {
	for _, supportedOS := range SupportedOSes() {
		if currentOS == supportedOS {
			return true
		}
	}
	return false
}
