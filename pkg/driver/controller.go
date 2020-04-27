package driver

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/ofek/csi-gcs/pkg/flags"
	"github.com/ofek/csi-gcs/pkg/util"
	"k8s.io/klog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cloud.google.com/go/storage"

	"google.golang.org/api/option"
)

func (d *GCSDriver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	klog.V(4).Infof("Method CreateVolume called with: %s", protosanitizer.StripSecrets(req))

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "missing name")
	}
	if len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing volume capabilities")
	}

	for _, capability := range req.GetVolumeCapabilities() {
		if capability.GetMount() != nil && capability.GetBlock() == nil {
			continue
		}
		return nil, status.Error(codes.InvalidArgument, "Only volumeMode Filesystem is supported")
	}

	// Default Options
	var options = map[string]string{
		"bucket":   util.BucketName(req.Name),
		"location": "US",
	}

	// Merge Secret Options
	options = flags.MergeSecret(options, req.Secrets)

	// Merge MountFlag Options
	for _, capability := range req.GetVolumeCapabilities() {
		options = flags.MergeMountOptions(options, capability.GetMount().GetMountFlags())
	}

	// Merge PVC Annotation Options
	pvcName, pvcNameSelected := req.Parameters["csi.storage.k8s.io/pvc/name"]
	pvcNamespace, pvcNamespaceSelected := req.Parameters["csi.storage.k8s.io/pvc/namespace"]

	var pvcAnnotations = map[string]string{}

	if pvcNameSelected && pvcNamespaceSelected {
		loadedPvcAnnotations, err := util.GetPvcAnnotations(pvcName, pvcNamespace)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to load PersistentVolumeClaim: %v", err)
		}

		pvcAnnotations = loadedPvcAnnotations
	}
	options = flags.MergeAnnotations(options, pvcAnnotations)

	// Merge Context
	if req.Parameters != nil {
		options = flags.MergeAnnotations(options, req.Parameters)
	}

	// Retrieve Key Secret
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

	// Check if Bucket Exists
	_, err = bucket.Attrs(ctx)
	if err == nil {
		klog.V(2).Infof("Bucket '%s' exists", options[flags.FLAG_BUCKET])
	} else {
		klog.V(2).Infof("Bucket '%s' does not exist, creating", options[flags.FLAG_BUCKET])

		projectId, projectIdExists := options[flags.FLAG_PROJECT_ID]
		if !projectIdExists {
			return nil, status.Errorf(codes.InvalidArgument, "Project Id not provided, bucket can't be created: %s", options[flags.FLAG_BUCKET])
		}

		if err := bucket.Create(ctx, projectId, &storage.BucketAttrs{Location: options[flags.FLAG_LOCATION]}); err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to create bucket: %v", err)
		}
	}

	// Get Capacity
	bucketAttrs, err := bucket.Attrs(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get bucket attrs: %v", err)
	}

	existingCapacity, err := util.BucketCapacity(bucketAttrs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get bucket capacity: %v", err)
	}

	// Check / Set Capacity
	newCapacity := int64(req.GetCapacityRange().GetRequiredBytes())
	if existingCapacity == 0 {
		_, err = util.SetBucketCapacity(ctx, bucket, newCapacity)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to set bucket capacity: %v", err)
		}
	} else if existingCapacity < newCapacity {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Volume with the same name: %s but with smaller size already exist", options[flags.FLAG_BUCKET]))
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      options[flags.FLAG_BUCKET],
			VolumeContext: options,
			CapacityBytes: newCapacity,
		},
	}, nil
}

func (d *GCSDriver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	klog.V(4).Infof("Method DeleteVolume called with: %s", protosanitizer.StripSecrets(req))

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing volume id")
	}

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
	bucket := client.Bucket(req.VolumeId)

	_, err = bucket.Attrs(ctx)
	if err == nil {
		if err := bucket.Delete(ctx); err != nil {
			return nil, status.Errorf(codes.Internal, "Error deleting bucket %s, %v", req.VolumeId, err)
		}
	} else {
		klog.V(2).Infof("Bucket '%s' does not exist, not deleting", req.VolumeId)
	}

	return &csi.DeleteVolumeResponse{}, nil
}

func (d *GCSDriver) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	klog.V(4).Infof("Method ControllerGetCapabilities called with: %s", protosanitizer.StripSecrets(req))

	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
		},
	}, nil
}

func (d *GCSDriver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	klog.V(4).Infof("Method ValidateVolumeCapabilities called with: %s", protosanitizer.StripSecrets(req))

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing volume id")
	}
	if len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing volume capabilities")
	}

	bucketName := req.VolumeId

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
	bucket := client.Bucket(bucketName)

	_, err = bucket.Attrs(ctx)

	if err != nil {
		return nil, status.Error(codes.NotFound, "volume does not exist")
	}

	for _, capability := range req.GetVolumeCapabilities() {
		if capability.GetMount() != nil && capability.GetBlock() == nil {
			continue
		}
		return &csi.ValidateVolumeCapabilitiesResponse{Message: "Only volumeMode Filesystem is supported"}, nil
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{

			VolumeContext:      req.GetVolumeContext(),
			VolumeCapabilities: req.GetVolumeCapabilities(),
			Parameters:         req.GetParameters(),
		},
	}, nil
}

func (d *GCSDriver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	klog.V(4).Infof("Method ControllerPublishVolume called with: %s", protosanitizer.StripSecrets(req))

	return nil, status.Error(codes.Unimplemented, "")
}

func (d *GCSDriver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	klog.V(4).Infof("Method ControllerUnpublishVolume called with: %s", protosanitizer.StripSecrets(req))

	return nil, status.Error(codes.Unimplemented, "")
}

func (d *GCSDriver) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	klog.V(4).Infof("Method GetCapacity called with: %s", protosanitizer.StripSecrets(req))

	return nil, status.Error(codes.Unimplemented, "")
}

func (d *GCSDriver) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	klog.V(4).Infof("Method ListVolumes called with: %s", protosanitizer.StripSecrets(req))

	return nil, status.Error(codes.Unimplemented, "")
}

func (d *GCSDriver) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	klog.V(4).Infof("Method CreateSnapshot called with: %s", protosanitizer.StripSecrets(req))

	return nil, status.Error(codes.Unimplemented, "")
}

func (d *GCSDriver) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	klog.V(4).Infof("Method DeleteSnapshot called with: %s", protosanitizer.StripSecrets(req))

	return nil, status.Error(codes.Unimplemented, "")
}

func (d *GCSDriver) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	klog.V(4).Infof("Method ListSnapshots called with: %s", protosanitizer.StripSecrets(req))

	return nil, status.Error(codes.Unimplemented, "")
}

func (d *GCSDriver) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	klog.V(4).Infof("Method ControllerExpandVolume called with: %s", protosanitizer.StripSecrets(req))

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing volume id")
	}

	// Retrieve Key Secret
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
	bucket := client.Bucket(req.VolumeId)

	// Check if Bucket Exists
	_, err = bucket.Attrs(ctx)
	if err == nil {
		klog.V(2).Infof("Bucket '%s' exists", req.VolumeId)
	} else {
		return nil, status.Errorf(codes.NotFound, "Bucket '%s' does not exist", req.VolumeId)
	}

	// Get Capacity
	bucketAttrs, err := bucket.Attrs(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get bucket attrs: %v", err)
	}

	existingCapacity, err := util.BucketCapacity(bucketAttrs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get bucket capacity: %v", err)
	}

	// Check / Set Capacity
	newCapacity := int64(req.GetCapacityRange().GetRequiredBytes())
	if newCapacity > existingCapacity {
		_, err = util.SetBucketCapacity(ctx, bucket, newCapacity)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to set bucket capacity: %v", err)
		}
	}

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         newCapacity,
		NodeExpansionRequired: false,
	}, nil
}
