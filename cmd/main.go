package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"k8s.io/klog"

	"github.com/ofek/csi-gcs/pkg/driver"
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
	setEnvVarFlags()
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

func setEnvVarFlags() {
	flagset := flag.CommandLine

	// I wish Golang had sets
	set := map[string]string{}

	// https://golang.org/pkg/flag/#FlagSet.Visit
	flagset.Visit(func(f *flag.Flag) {
		set[f.Name] = ""
	})

	// https://golang.org/pkg/flag/#FlagSet.VisitAll
	flagset.VisitAll(func(f *flag.Flag) {
		envVar := strings.Replace(strings.ToUpper(f.Name), "-", "_", -1)

		if val := os.Getenv(envVar); val != "" {
			if _, defined := set[f.Name]; !defined {
				_ = flagset.Set(f.Name, val)
			}
		}

		// Display it in the help text too
		f.Usage = fmt.Sprintf("%s [%s]", f.Usage, envVar)
	})
}
