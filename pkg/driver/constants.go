package driver

const (
	CSIDriverName   = "gcs.csi.ofek.dev"
	BucketMountPath = "/var/lib/kubelet/pods"
	KeyStoragePath  = "/tmp/keys"
	DefaultGid      = 63147
	DefaultDirMode  = 0775
	DefaultFileMode = 0664
)
