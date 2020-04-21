package driver

import (
	"context"
	"errors"
	"net"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"k8s.io/klog"

	"github.com/ofek/csi-gcs/pkg/util"
)

type GCSDriver struct {
	name       string
	nodeName   string
	endpoint   string
	mountPoint string
	version    string
	server     *grpc.Server
}

func NewGCSDriver(name, node, endpoint string) (*GCSDriver, error) {
	return &GCSDriver{
		name:       name,
		nodeName:   node,
		endpoint:   endpoint,
		mountPoint: BucketMountPath,
		version:    driverVersion,
	}, nil
}

func (d *GCSDriver) Run() error {
	if len(d.mountPoint) == 0 {
		return errors.New("--bucket-mount-path is required")
	}

	scheme, address, err := util.ParseEndpoint(d.endpoint)
	if err != nil {
		return err
	}

	listener, err := net.Listen(scheme, address)
	if err != nil {
		return err
	}

	logHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			klog.V(4).Infof("Method %s completed", info.FullMethod)
		} else {
			klog.Errorf("Method %s failed with error: %v", info.FullMethod, err)
		}
		return resp, err
	}

	klog.V(1).Infof("Starting Google Cloud Storage CSI Driver - driver: `%s`, version: `%s`, gRPC socket: `%s`", d.name, d.version, d.endpoint)
	d.server = grpc.NewServer(grpc.UnaryInterceptor(logHandler))
	csi.RegisterIdentityServer(d.server, d)
	csi.RegisterNodeServer(d.server, d)
	csi.RegisterControllerServer(d.server, d)
	return d.server.Serve(listener)
}

func (d *GCSDriver) stop() {
	d.server.Stop()
	klog.V(1).Info("CSI driver stopped")
}
