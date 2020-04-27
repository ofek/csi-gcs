package driver

import (
	"context"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"

	"github.com/ofek/csi-gcs/pkg/flags"
	"github.com/ofek/csi-gcs/pkg/util"
)

func (driver *GCSDriver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.V(4).Infof("Method NodePublishVolume called with: %s", protosanitizer.StripSecrets(req))

	if req.GetVolumeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume Volume Capability must be provided")
	}

	if req.VolumeCapability.GetMount() == nil || req.VolumeCapability.GetBlock() != nil {
		return nil, status.Error(codes.InvalidArgument, "Only volumeMode Filesystem is supported")
	}

	// Default Options
	var options = map[string]string{
		"bucket":   req.GetVolumeId(),
		"gid":      strconv.FormatInt(DefaultGid, 10),
		"dirMode":  "0" + strconv.FormatInt(DefaultDirMode, 8),
		"fileMode": "0" + strconv.FormatInt(DefaultFileMode, 8),
	}

	// Merge Secret Options
	options = flags.MergeSecret(options, req.Secrets)

	// Merge MountFlag Options
	options = flags.MergeMountOptions(options, req.GetVolumeCapability().GetMount().GetMountFlags())

	// Merge Volume Context
	if req.VolumeContext != nil {
		options = flags.MergeFlags(options, req.VolumeContext)
	}

	// Retrieve Secret Key
	keyFile, err := util.GetKey(req.Secrets, KeyStoragePath)
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
	bucket := client.Bucket(options[flags.FLAG_BUCKET])

	bucketExists, err := util.BucketExists(ctx, bucket)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check if bucket exists: %v", err)
	}
	if !bucketExists {
		return nil, status.Errorf(codes.NotFound, "Bucket %s does not exist", options[flags.FLAG_BUCKET])
	}

	notMnt, err := CheckMount(req.TargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	mounter, err := NewGcsFuseMounter(options[flags.FLAG_BUCKET], keyFile, flags.ExtraFlags(options))
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

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	notMount, err := CheckMount(req.GetTargetPath())
	if err != nil || notMount {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	if err := FuseUnmount(req.GetTargetPath()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.V(4).Infof("bucket %s has been unmounted.", req.GetVolumeId())

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (driver *GCSDriver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	klog.V(4).Infof("Method NodeGetInfo called with: %s", protosanitizer.StripSecrets(req))

	return &csi.NodeGetInfoResponse{NodeId: driver.nodeName}, nil
}

func (driver *GCSDriver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	klog.V(4).Infof("Method NodeGetCapabilities called with: %s", protosanitizer.StripSecrets(req))

	return &csi.NodeGetCapabilitiesResponse{Capabilities: []*csi.NodeServiceCapability{
		{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
				},
			},
		},
	}}, nil
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

	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetVolumePath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume path missing in request")
	}

	notMount, err := CheckMount(req.GetVolumePath())
	if err != nil || notMount {
		return nil, status.Error(codes.NotFound, "Volume is not mounted")
	}

	return &csi.NodeExpandVolumeResponse{}, nil
}
