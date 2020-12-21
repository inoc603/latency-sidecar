package main

import (
	"fmt"
	"os/exec"
	"time"
)

// TcLatencySetter implements LatencySetter by calling the tc command.
type TcLatencySetter struct{}

// SetLatency sets extra latency of the given interface
func (t TcLatencySetter) SetLatency(iname string, latency time.Duration) error {
	var cmd *exec.Cmd

	if latency <= 0 {
		cmd = exec.Command(
			"tc",
			"qdisc",
			"del",
			"dev",
			iname,
			"root",
		)
	} else {
		cmd = exec.Command(
			"tc",
			"qdisc",
			"add",
			"dev",
			iname,
			"root",
			"netem",
			"delay",
			latency.String(),
		)
	}

	fmt.Println(cmd.String())

	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil {
		return fmt.Errorf("command %s failed: %w", cmd.String(), err)
	}

	return nil
}
