package V2

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/containerd/cgroups"
)

// v2MountPoint returns the mount point where the cgroup
// mountpoints are mounted in a single hiearchy
func v2MountPoint() (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var (
			text      = scanner.Text()
			fields    = strings.Split(text, " ")
			numFields = len(fields)
		)
		if numFields < 10 {
			return "", fmt.Errorf("mountinfo: bad entry %q", text)
		}
		if fields[numFields-3] == "cgroup2" {
			return fields[4], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", cgroups.ErrMountPointNotExist
}
