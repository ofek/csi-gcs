package sanity_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/kubernetes-csi/csi-test/v3/pkg/sanity"
	"github.com/ofek/csi-gcs/pkg/driver"
	"k8s.io/klog"
)

func TestCsiGcs(t *testing.T) {
	endpointFile, err := ioutil.TempFile("", "csi-gcs.*.sock")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(endpointFile.Name())

	stagingPath, err := ioutil.TempDir("", "csi-gcs-staging")
	if err != nil {
		t.Fatal(err)
	}

	os.Remove(stagingPath)

	targetPath, err := ioutil.TempDir("", "csi-gcs-target")
	if err != nil {
		t.Fatal(err)
	}

	os.Remove(targetPath)

	var endpoint = "unix://"
	endpoint += endpointFile.Name()

	d, err := driver.NewGCSDriver(driver.CSIDriverName, "test-node", endpoint)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	go func() {
		if err = d.Run(); err != nil {
			t.Fatal(err)
		}
	}()

	config := sanity.NewTestConfig()
	// Set configuration options as needed
	config.Address = endpoint
	config.SecretsFile = "./secret.yaml"
	config.StagingPath = stagingPath
	config.TargetPath = targetPath
	config.RemoveTargetPath = func(patargetPathth string) error {
		return os.RemoveAll(targetPath)
	}

	// Now call the test suite
	sanity.Test(t, config)
}
