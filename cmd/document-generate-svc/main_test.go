package main

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/kinneko-de/restaurant-document-generate-svc/internal/app/operation/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain_MetricConfigIsMissing(t *testing.T) {
	if os.Getenv("EXECUTE") == "1" {
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_MetricConfigIsMissing")
	cmd.Env = append(os.Environ(), "EXECUTE=1")
	err := cmd.Run()
	require.NotNil(t, err)
	exitCode := err.(*exec.ExitError).ExitCode()
	assert.Equal(t, 40, exitCode)
}

// test does not run on windows
// In case you broke something, the test will run forever
// In the pipeline you will see:
// panic: test timed out after 5m0s
// running tests:
// TestMain_ApplicationListenToInterrupt_GracefullShutdown (5m0s)
func TestMain_ApplicationListenToSIGTERM_AndGracefullyShutdown(t *testing.T) {
	if os.Getenv("EXECUTE") == "1" {
		main()
		return
	}

	t.Setenv(metric.OtelMetricEndpointEnv, "http://localhost")
	t.Setenv(metric.ServiceNameEnv, "blub")
	cmd := exec.Command(os.Args[0], "-test.run=TestMain_ApplicationListenToSIGTERM_AndGracefullyShutdown")
	cmd.Env = append(os.Environ(), "EXECUTE=1")
	err := cmd.Start()
	require.Nil(t, err)
	time.Sleep(1 * time.Second)
	cmd.Process.Signal(syscall.SIGTERM)
	err = cmd.Wait()
	require.Nil(t, err)
	exitCode := cmd.ProcessState.ExitCode()
	assert.Equal(t, 0, exitCode)
}

// test does not run on windows
// In case you broke something, the test will run forever
// In the pipeline you will see:
// panic: test timed out after 5m0s
// running tests:
// TestMain_ApplicationListenToInterrupt_GracefullShutdown (5m0s)
func TestMain_ProcessAlreadyListenToPort_AppCrash(t *testing.T) {
	if os.Getenv("EXECUTE") == "1" {
		main()
		return
	}

	t.Setenv(metric.OtelMetricEndpointEnv, "http://localhost")
	t.Setenv(metric.ServiceNameEnv, "blub")
	blockingcmd := exec.Command(os.Args[0], "-test.run=TestMain_ProcessAlreadyListenToPort_AppCrash")
	blockingcmd.Env = append(os.Environ(), "EXECUTE=1")
	blockingErr := blockingcmd.Start()
	require.Nil(t, blockingErr)
	time.Sleep(1 * time.Second) // give the service some time to start
	cmd := exec.Command(os.Args[0], "-test.run=TestMain_ProcessAlreadyListenToPort_AppCrash")
	cmd.Env = append(os.Environ(), "EXECUTE=1")
	err := cmd.Run()
	require.NotNil(t, err)
	exitCode := err.(*exec.ExitError).ExitCode()
	assert.Equal(t, 50, exitCode)
	blockingcmd.Process.Kill()
}
