package driver

import (
	"context"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"

	"github.com/ofek/csi-gcs/pkg/util"
)

func (driver *GCSDriver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.V(4).Infof("Method NodePublishVolume called with: %s", protosanitizer.StripSecrets(req))

	volumeId := req.GetVolumeId()
	if volumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	volumePath := req.TargetPath
	if volumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Volume Capability must be provided")
	}

	options := req.VolumeContext

	var bucketName string
	if contextBucket, contextBucketSelected := options["bucket"]; contextBucketSelected {
		bucketName = contextBucket
	} else if secretBucket, secretBucketSelected := req.Secrets["bucket"]; secretBucketSelected {
		bucketName = secretBucket
	} else {
		klog.V(2).Infof("A bucket was not selected, defaulting to the volume name %s", volumeId)
		bucketName = volumeId
	}

	keyFile, err := util.GetKey(req.Secrets, options, KeyStoragePath)
	if err != nil {
		return nil, err
	}
	defer util.CleanupKey(keyFile, KeyStoragePath)

	// Creates a client.
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(keyFile))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create client: %v", err)
	}

	// Creates a Bucket instance.
	bucket := client.Bucket(bucketName)

	bucketExists, err := util.BucketExists(ctx, bucket)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check if bucket exists: %v", err)
	}
	if !bucketExists {
		return nil, status.Errorf(codes.NotFound, "Bucket %s does not exist", bucketName)
	}

	var extraFlags = []string{}
	if flags, exists := options["flags"]; exists {
		parsedFlags := strings.Fields(flags)
		if parsedFlags != nil {
			extraFlags = append(extraFlags, parsedFlags...)
		}
	}

	notMnt, err := CheckMount(req.TargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	mounter, err := NewGcsFuseMounter(bucketName, keyFile, extraFlags)
	if err != nil {
		return nil, err
	}
	if err := mounter.Mount(req.TargetPath); err != nil {
		return nil, err
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (driver *GCSDriver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (response *csi.NodeUnpublishVolumeResponse, err error) {
	klog.V(4).Infof("Method NodeUnpublishVolume called with: %s", protosanitizer.StripSecrets(req))

	volumeId := req.GetVolumeId()
	targetPath := req.GetTargetPath()

	// Check arguments
	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	notMount, err := CheckMount(targetPath)
	if err != nil || notMount {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	if err := FuseUnmount(targetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.V(4).Infof("bucket %s has been unmounted.", volumeId)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (driver *GCSDriver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	klog.V(4).Infof("Method NodeGetInfo called with: %s", protosanitizer.StripSecrets(req))

	return &csi.NodeGetInfoResponse{NodeId: driver.nodeName}, nil
}

func (driver *GCSDriver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	klog.V(4).Infof("Method NodeGetCapabilities called with: %s", protosanitizer.StripSecrets(req))

	return &csi.NodeGetCapabilitiesResponse{Capabilities: []*csi.NodeServiceCapability{}}, nil
}

func (driver *GCSDriver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	klog.V(4).Infof("Method NodeStageVolume called with: %s", protosanitizer.StripSecrets(req))

	return nil, status.Errorf(codes.Unimplemented, "NodeStageVolume: not implemented by %s", driver.name)
}

func (driver *GCSDriver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	klog.V(4).Infof("Method NodeUnstageVolume called with: %s", protosanitizer.StripSecrets(req))

	return nil, status.Errorf(codes.Unimplemented, "NodeUnstageVolume: not implemented by %s", driver.name)
}

func (driver *GCSDriver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	klog.V(4).Infof("Method NodeGetVolumeStats called with: %s", protosanitizer.StripSecrets(req))

	return nil, status.Errorf(codes.Unimplemented, "NodeGetVolumeStats: not implemented by %s", driver.name)
}

func (driver *GCSDriver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	klog.V(4).Infof("Method NodeExpandVolume called with: %s", protosanitizer.StripSecrets(req))

	return nil, status.Errorf(codes.Unimplemented, "NodeExpandVolume: not implemented by %s", driver.name)
}
