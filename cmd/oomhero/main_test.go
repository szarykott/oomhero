package main

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/ricardomaraschini/oomhero/proc"
	"github.com/stretchr/testify/assert"
)

var (
	testTicker = make(chan time.Time)
)

func TestNoOp(t *testing.T) {
	t.Cleanup(resetState)
	p := newTestProcess(1)

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	tickXTimes(3)

	assert.Equal(t, 0, len(p.receivedSignals))
}

func TestSingleWarningReceivedDuringCooldown(t *testing.T) {
	t.Cleanup(resetState)

	cooldown = 60

	p := newTestProcess(1)
	p.memoryUsage = warning + 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	tickXTimes(3)

	assert.Equal(t, 1, len(p.receivedSignals))
	assert.Equal(t, syscall.SIGUSR1, p.receivedSignals[0])
}

func TestNextWarningReceivedAsCooldownElapses(t *testing.T) {
	t.Cleanup(resetState)

	cooldown = 60

	p := newTestProcess(1)
	p.memoryUsage = warning + 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	tickXTimes(int(cooldown) + 1)

	assert.Equal(t, 2, len(p.receivedSignals))
	assert.Equal(t, syscall.SIGUSR1, p.receivedSignals[0])
	assert.Equal(t, syscall.SIGUSR1, p.receivedSignals[1])
}

func TestMultipleWarningReceivedWithDefaultCooldown(t *testing.T) {
	t.Cleanup(resetState)

	p := newTestProcess(1)
	p.memoryUsage = warning + 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	tickXTimes(3)

	assert.Greater(t, len(p.receivedSignals), 1)
	assert.Equal(t, syscall.SIGUSR1, p.receivedSignals[0])
}

func TestSingleCriticalReceivedDuringCooldown(t *testing.T) {
	t.Cleanup(resetState)

	cooldown = 60

	p := newTestProcess(1)
	p.memoryUsage = critical + 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	tickXTimes(3)

	assert.Equal(t, 1, len(p.receivedSignals))
	assert.Equal(t, syscall.SIGUSR2, p.receivedSignals[0])
}

func TestNextCriticalReceivedAsCooldownElapses(t *testing.T) {
	t.Cleanup(resetState)

	cooldown = 60

	p := newTestProcess(1)
	p.memoryUsage = critical + 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	tickXTimes(int(cooldown) + 1)

	assert.Equal(t, 2, len(p.receivedSignals))
	assert.Equal(t, syscall.SIGUSR2, p.receivedSignals[0])
	assert.Equal(t, syscall.SIGUSR2, p.receivedSignals[1])
}

func TestMultipleCriticalReceivedWithDefaultCooldown(t *testing.T) {
	t.Cleanup(resetState)

	p := newTestProcess(1)
	p.memoryUsage = critical + 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	tickXTimes(3)

	assert.Greater(t, len(p.receivedSignals), 1)
	assert.Equal(t, syscall.SIGUSR2, p.receivedSignals[0])
}

func TestSingleCriticalAndWarningReceivedAsMemoryUsageGrowsDuringCooldown(t *testing.T) {
	t.Cleanup(resetState)

	cooldown = 60

	p := newTestProcess(1)
	p.memoryUsage = 0

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	p.memoryUsage = warning + 1

	tickXTimes(2)

	p.memoryUsage = critical + 1

	tickXTimes(2)

	assert.Equal(t, 2, len(p.receivedSignals))
	assert.Equal(t, syscall.SIGUSR1, p.receivedSignals[0])
	assert.Equal(t, syscall.SIGUSR2, p.receivedSignals[1])
}

func TestSingleWarningReceivedWhenMemoryUsageOscilatesDuringCooldown(t *testing.T) {
	t.Cleanup(resetState)

	cooldown = 60

	p := newTestProcess(1)
	p.memoryUsage = warning - 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	p.memoryUsage = warning + 1

	tickXTimes(2)

	p.memoryUsage = warning - 1

	tickXTimes(2)

	p.memoryUsage = warning + 1

	tickXTimes(2)

	assert.Equal(t, 1, len(p.receivedSignals))
	assert.Equal(t, syscall.SIGUSR1, p.receivedSignals[0])
}

func TestMultipleWarningReceivedWhenMemoryUsageOscilatesWithDefaultCooldown(t *testing.T) {
	t.Cleanup(resetState)

	p := newTestProcess(1)
	p.memoryUsage = warning - 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	p.memoryUsage = warning + 1

	tickXTimes(2)

	p.memoryUsage = warning - 1

	tickXTimes(2)

	p.memoryUsage = warning + 1

	tickXTimes(2)

	assert.Greater(t, len(p.receivedSignals), 1)
	assert.Equal(t, syscall.SIGUSR1, p.receivedSignals[0])
}

func TestSingleCriticalAndWarningReceivedWhenMemoryUsageOscilatesDuringCooldown(t *testing.T) {
	t.Cleanup(resetState)

	cooldown = 60

	p := newTestProcess(1)
	p.memoryUsage = critical - 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	p.memoryUsage = critical + 1

	tickXTimes(2)

	p.memoryUsage = critical - 1

	tickXTimes(2)

	p.memoryUsage = critical + 1

	tickXTimes(2)

	assert.Equal(t, 2, len(p.receivedSignals))
	assert.Equal(t, syscall.SIGUSR2, p.receivedSignals[0])
	assert.Equal(t, syscall.SIGUSR1, p.receivedSignals[1])
}

func TestMultipleCriticalAndWarningReceivedWhenMemoryUsageOscilatesWithDefaultCooldown(t *testing.T) {
	t.Cleanup(resetState)

	p := newTestProcess(1)
	p.memoryUsage = critical - 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	p.memoryUsage = critical + 1

	tickXTimes(2)

	p.memoryUsage = critical - 1

	tickXTimes(2)

	p.memoryUsage = critical + 1

	tickXTimes(2)

	assert.Greater(t, len(p.receivedSignals), 2)
	assert.Contains(t, p.receivedSignals, syscall.SIGUSR1)
	assert.Contains(t, p.receivedSignals, syscall.SIGUSR2)
}

func TestWarningSignalEnvSettingIsRespected(t *testing.T) {
	t.Cleanup(resetState)
	t.Setenv("WARNING_SIGNAL", "SIGTERM")

	cooldown = 60

	p := newTestProcess(1)
	p.memoryUsage = warning + 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	tickXTimes(3)

	assert.Equal(t, 1, len(p.receivedSignals))
	assert.Equal(t, syscall.SIGTERM, p.receivedSignals[0])
}

func TestCriticalSignalEnvSettingIsRespected(t *testing.T) {
	t.Cleanup(resetState)
	t.Setenv("CRITICAL_SIGNAL", "SIGTERM")

	cooldown = 60

	p := newTestProcess(1)
	p.memoryUsage = critical + 1

	ps := TestProcesses{
		items: []proc.Process{&p},
	}

	go watchProcesses(testTicker, ps.getProcesses)

	tickXTimes(3)

	assert.Equal(t, 1, len(p.receivedSignals))
	assert.Equal(t, syscall.SIGTERM, p.receivedSignals[0])
}

func resetState() {
	cooldown = 1
	close(testTicker)
	testTicker = make(chan time.Time)
}

func tickXTimes(n int) {
	now := time.Now()
	for i := 0; i < n; i++ {
		testTicker <- now
		time.Sleep(50 * time.Millisecond) // let the other gorutine do its things
		now = now.Add(time.Second)
	}
}

type TestProcesses struct {
	items []proc.Process
}

func (p TestProcesses) getProcesses() ([]proc.Process, error) {
	return p.items, nil
}

type TestProcess struct {
	pid             int
	receivedSignals []os.Signal
	memoryUsage     uint64
}

func newTestProcess(pid int) TestProcess {
	return TestProcess{
		pid:             pid,
		receivedSignals: make([]os.Signal, 0),
		memoryUsage:     0,
	}
}

func (p *TestProcess) Pid() int {
	return p.pid
}

func (p *TestProcess) Signal(s os.Signal) error {
	p.receivedSignals = append(p.receivedSignals, s)

	return nil
}

func (p *TestProcess) MemoryUsagePercent() (uint64, error) {
	return p.memoryUsage, nil
}
