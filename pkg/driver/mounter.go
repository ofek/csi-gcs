package driver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	"github.com/mitchellh/go-ps"
	"github.com/ofek/csi-gcs/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"k8s.io/utils/mount"
)

// Implements Mounter
type GcsFuseMounter struct {
	bucket  string
	keyFile string
	flags   []string
}

const (
	gcsFuseCommand = "gcsfuse"
)

func NewGcsFuseMounter(bucket string, keyFile string, flags []string) (*GcsFuseMounter, error) {
	return &GcsFuseMounter{
		bucket:  bucket,
		keyFile: keyFile,
		flags:   flags,
	}, nil
}

func (gcsfuse *GcsFuseMounter) Mount(target string) error {
	args := []string{fmt.Sprintf("--key-file=%s", gcsfuse.keyFile), "-o=allow_other", "--foreground"}
	args = append(args, gcsfuse.flags...)
	args = append(args, gcsfuse.bucket, target)

	if err := util.CreateDir(target); err != nil {
		return status.Errorf(codes.Internal, "Unable to create target path %s: %v", target, err)
	}

	return FuseMount(target, gcsFuseCommand, args, gcsfuse.bucket)
}

func FuseMount(path string, command string, args []string, bucket string) error {
	cmd := exec.Command(command, args...)
	klog.V(3).Infof("Mounting fuse with command: %s and args: %s", command, args)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	scannerStdout := bufio.NewScanner(stdout)
	go func() {
		for scannerStdout.Scan() {
			text := scannerStdout.Text()
			klog.V(3).Infof("[%s] %s", bucket, text)
		}
	}()

	scannerStderr := bufio.NewScanner(stderr)
	go func() {
		for scannerStderr.Scan() {
			text := scannerStderr.Text()
			klog.Errorf("[%s] %s", bucket, text)
		}
	}()

	if err := cmd.Start(); err != nil {
		return err
	}

	return WaitForMount(path, 10*time.Second)
}

func FuseUnmount(path string) error {
	if err := mount.New("").Unmount(path); err != nil {
		return err
	}
	// as fuse quits immediately, we will try to wait until the process is done
	process, err := FindFuseMountProcess(path)
	if err != nil {
		klog.Errorf("Error getting PID of fuse mount: %s", err)
		return nil
	}
	if process == nil {
		klog.Warningf("Unable to find PID of fuse mount %s, it must have finished already", path)
		return nil
	}
	klog.Infof("Found fuse pid %v of mount %s, checking if it still runs", process.Pid, path)
	return WaitForProcess(process, 1)
}

func SetBucketCapacity(ctx context.Context, bucket *storage.BucketHandle, capacity int64) (attrs *storage.BucketAttrs, err error) {
	var uattrs = storage.BucketAttrsToUpdate{}

	uattrs.SetLabel("capacity", strconv.FormatInt(capacity, 10))

	return bucket.Update(ctx, uattrs)
}

func FindFuseMountProcess(path string) (*os.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range processes {
		cmdLine, err := GetCmdLine(p.Pid())
		if err != nil {
			klog.Errorf("Unable to get cmdline of PID %v: %s", p.Pid(), err)
			continue
		}
		if strings.Contains(cmdLine, path) {
			klog.Infof("Found matching pid %v on path %s", p.Pid(), path)
			return os.FindProcess(p.Pid())
		}
	}
	return nil, nil
}

func GetCmdLine(pid int) (string, error) {
	cmdLineFile := fmt.Sprintf("/proc/%v/cmdline", pid)
	cmdLine, err := ioutil.ReadFile(cmdLineFile)
	if err != nil {
		return "", err
	}
	return string(cmdLine), nil
}

func WaitForProcess(p *os.Process, backoff int) error {
	if backoff == 20 {
		return fmt.Errorf("Timeout waiting for PID %v to end", p.Pid)
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		klog.Warningf("Fuse process does not seem active or we are unprivileged: %s", err)
		return nil
	}
	klog.Infof("Fuse process with PID %v still active, waiting...", p.Pid)
	time.Sleep(time.Duration(backoff*100) * time.Millisecond)
	return WaitForProcess(p, backoff+1)
}

func WaitForMount(path string, timeout time.Duration) error {
	var elapsed time.Duration
	var interval = 10 * time.Millisecond
	for {
		notMount, err := mount.IsNotMountPoint(mount.New(""), path)
		if err != nil {
			return err
		}
		if !notMount {
			return nil
		}
		time.Sleep(interval)
		elapsed = elapsed + interval
		if elapsed >= timeout {
			return errors.New("Timeout waiting for mount")
		}
	}
}

func CheckMount(targetPath string) (bool, error) {
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return false, err
			}
			notMnt = true
		} else {
			return false, err
		}
	}
	return notMnt, nil
}
