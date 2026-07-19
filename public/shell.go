package public

import (
	alog "agent/logger"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func ExecShell(commandLine string) (error, string) {
	alog.Log.Println("ExecShell:", commandLine)
	fields := strings.Fields(commandLine)
	if len(fields) == 0 {
		return fmt.Errorf("empty command"), ""
	}

	return runCommand(30*time.Second, fields[0], fields[1:]...)
}

func ExecWgetEn(url string) (error, string) {
	alog.Log.Println("ExecWgetEn:", url)
	return runCommand(60*time.Second, "wget", "-P", "/tmp", url)
}

func runCommand(timeout time.Duration, name string, args ...string) (error, string) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		return err, out.String()
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		return fmt.Errorf("command timed out"), out.String()
	case err := <-done:
		return err, out.String()
	}
}
