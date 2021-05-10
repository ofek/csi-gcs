package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ofek/csi-gcs/pkg/driver"
	"github.com/ofek/csi-gcs/pkg/util"
	"k8s.io/klog"
)

var (
	version            = "development"
	nodeNameFlag       = flag.String("node-name", "", "Node identifier")
	driverNameFlag     = flag.String("driver-name", driver.CSIDriverName, "CSI driver name")
	endpointFlag       = flag.String("csi-endpoint", "unix:///csi/csi.sock", "CSI endpoint")
	versionFlag        = flag.Bool("version", false, "Print the version and exit")
	deleteOrphanedPods = flag.Bool("delete-orphaned-pods", false, "Delete Orphaned Pods on StartUp")
)

func main() {
	_ = flag.Set("alsologtostderr", "true")
	klog.InitFlags(nil)
	util.SetEnvVarFlags()
	flag.Parse()

	if *versionFlag {
		versionJSON, err := driver.GetVersionJSON()
		if err != nil {
			klog.Error(err.Error())
			os.Exit(1)
		}
		fmt.Println(versionJSON)
		os.Exit(0)
	}

	d, err := driver.NewGCSDriver(*driverNameFlag, *nodeNameFlag, *endpointFlag, version, *deleteOrphanedPods)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	if err = d.Run(); err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}
}
