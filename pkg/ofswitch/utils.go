package ofswitch

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func getOFPortByLinkName(name string) (int, error) {
	cmd := exec.Command("ovs-vsctl", "--columns=name,ofport", "list", "interface", name)
	out, err := cmd.Output()
	if err != nil {
		return -1, err
	}

	outs := strings.Split(string(out), "\n")
	nameCol := outs[0]
	portCol := outs[1]

	if !strings.Contains(nameCol, "name") {
		return -1, fmt.Errorf("Failed to parse name column %v", nameCol)
	}
	nameRes := strings.Trim(strings.Trim(strings.Split(nameCol, ":")[1], " "), "\"")

	if !strings.Contains(portCol, "ofport") {
		return -1, fmt.Errorf("Failed to parse ofport column %v", portCol)
	}
	portStr := strings.Trim(strings.Split(portCol, ":")[1], " ")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return -1, err
	}

	if string(nameRes) != string(name) {
		return -1, fmt.Errorf("Failed expected:%v actual:%v", name, nameRes)
	}

	return port, nil
}
