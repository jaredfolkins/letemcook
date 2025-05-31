package tests

import (
	"os/exec"
	"sync"
	"syscall"
	"time"
)

var (
	processRegistryMutex sync.RWMutex
	testProcesses        = make(map[int]*exec.Cmd) // map of PID to command
	testServerPorts      = make(map[int]bool)      // set of ports used by test servers
)

// RegisterTestProcess adds a process to the test registry for safe cleanup
func RegisterTestProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	processRegistryMutex.Lock()
	defer processRegistryMutex.Unlock()

	testProcesses[cmd.Process.Pid] = cmd
}

// UnregisterTestProcess removes a process from the test registry
func UnregisterTestProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	processRegistryMutex.Lock()
	defer processRegistryMutex.Unlock()

	delete(testProcesses, cmd.Process.Pid)
}

// RegisterTestPort adds a port to the test registry
func RegisterTestPort(port int) {
	processRegistryMutex.Lock()
	defer processRegistryMutex.Unlock()

	testServerPorts[port] = true
}

// UnregisterTestPort removes a port from the test registry
func UnregisterTestPort(port int) {
	processRegistryMutex.Lock()
	defer processRegistryMutex.Unlock()

	delete(testServerPorts, port)
}

// GetTestProcesses returns a copy of all registered test processes
func GetTestProcesses() map[int]*exec.Cmd {
	processRegistryMutex.RLock()
	defer processRegistryMutex.RUnlock()

	result := make(map[int]*exec.Cmd)
	for pid, cmd := range testProcesses {
		result[pid] = cmd
	}
	return result
}

// GetTestPorts returns a copy of all registered test ports
func GetTestPorts() []int {
	processRegistryMutex.RLock()
	defer processRegistryMutex.RUnlock()

	var result []int
	for port := range testServerPorts {
		result = append(result, port)
	}
	return result
}

// CleanupRegisteredProcesses safely kills only the processes we've registered
func CleanupRegisteredProcesses() {
	processes := GetTestProcesses()

	for _, cmd := range processes {
		if cmd != nil && cmd.Process != nil {
			// Try graceful shutdown first
			cmd.Process.Signal(syscall.SIGTERM)
		}
	}

	// Give processes time to shut down gracefully
	time.Sleep(1 * time.Second)

	// Force kill any remaining processes
	for _, cmd := range processes {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
		}
		UnregisterTestProcess(cmd)
	}
}
