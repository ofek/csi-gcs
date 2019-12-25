package driver

import (
	"encoding/json"
	"fmt"
	"runtime"
)

// These will be set at build time.
var (
	driverVersion string
)

type VersionInfo struct {
	DriverVersion string `json:"driverVersion"`
	GoVersion     string `json:"goVersion"`
	Compiler      string `json:"compiler"`
	Platform      string `json:"platform"`
}

func GetVersion() VersionInfo {
	return VersionInfo{
		DriverVersion: driverVersion,
		GoVersion:     runtime.Version(),
		Compiler:      runtime.Compiler,
		Platform:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func GetVersionJSON() (string, error) {
	versionInfo := GetVersion()

	marshalled, err := json.MarshalIndent(&versionInfo, "", "  ")
	if err != nil {
		return "", err
	}

	return string(marshalled), nil
}
