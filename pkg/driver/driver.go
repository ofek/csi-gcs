package driver

import (
	"context"
	"errors"
	"net"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"k8s.io/klog"

	"github.com/ofek/csi-gcs/pkg/util"

	"k8s.io/utils/mount"
)

type GCSDriver struct {
	name               string
	nodeName           string
	endpoint           string
	mountPoint         string
	version            string
	server             *grpc.Server
	mounter            mount.Interface
	deleteOrphanedPods bool
}

func NewGCSDriver(name, node, endpoint string, version string, deleteOrphanedPods bool) (*GCSDriver, error) {
	return &GCSDriver{
		name:               name,
		nodeName:           node,
		endpoint:           endpoint,
		mountPoint:         BucketMountPath,
		version:            version,
		mounter:            mount.New(""),
		deleteOrphanedPods: deleteOrphanedPods,
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

	if d.deleteOrphanedPods {
		err = d.RunPodCleanup()

		if err != nil {
			klog.Errorf("RunPodCleanup failed with error: %v", err)
		}
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

func (d *GCSDriver) RunPodCleanup() (err error) {
	publishedVolumes, err := util.GetRegisteredMounts(d.nodeName)
	if err != nil {
		return err
	}

	for _, publishedVolume := range publishedVolumes.Items {
		// Killing Pod because its Volume is no longer mounted
		err = util.DeletePod(publishedVolume.Spec.Pod.Namespace, publishedVolume.Spec.Pod.Name)
		if err == nil {
			klog.V(4).Infof("Deleted Pod %s/%s bacause its volume was no longer mounted", publishedVolume.Spec.Pod.Namespace, publishedVolume.Spec.Pod.Name)
		} else {
			klog.Errorf("Could not delete pod %s/%s because it was no longer mounted because of error: %v", publishedVolume.Spec.Pod.Namespace, publishedVolume.Spec.Pod.Name, err)
		}
	}

	return nil
}
