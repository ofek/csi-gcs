package driver

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"

	"github.com/ofek/csi-gcs/pkg/util"
)

func (d *GCSDriver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	klog.V(4).Infof("Method NodePublishVolume called with: %s", protosanitizer.StripSecrets(req))

	d.m.Lock()
	defer d.m.Unlock()

	volumeId := req.GetVolumeId()
	if volumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	} else if v, exists := d.volumes[volumeId]; exists {
		v.count++
		klog.V(2).Infof("Volume %s already exists, now there are %d references", volumeId, v.count)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	volumePath := req.TargetPath
	if volumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	options := req.VolumeContext

	var bucket string
	if secretBucket, secretBucketSelected := req.Secrets["bucket"]; secretBucketSelected {
		bucket = secretBucket
	} else if contextBucket, contextBucketSelected := options["bucket"]; contextBucketSelected {
		bucket = contextBucket
	} else {
		klog.V(2).Infof("A bucket was not selected, defaulting to the volume name %s", volumeId)
		bucket = volumeId
	}

	keyFile, keyFileSelected := options["key_file"]
	if keyFileSelected {
		klog.V(2).Infof("Using service account key located at %s", keyFile)
	} else {
		keyName, keyNameSelected := options["key_name"]
		if !keyNameSelected {
			keyName = "key"
		}
		klog.V(2).Infof("Using secret name '%s' as the service account key", keyName)

		keyContents, keyNameExists := req.Secrets[keyName]
		if !keyNameExists {
			return nil, status.Errorf(codes.Internal, "Secret '%s' is unavailable", keyName)
		}

		klog.V(5).Info("Saving key contents to a temporary location")
		tempKeyFile, err := util.CreateFile(KeyStoragePath, keyContents)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Unable to save secret '%s' to %s", keyName, KeyStoragePath)
		}
		keyFile = tempKeyFile
	}

	allFlags := []string{fmt.Sprintf("--key-file=%s", keyFile), "-o=allow_other"}
	if flags, exists := options["flags"]; exists {
		parsedFlags := strings.Fields(flags)
		if parsedFlags != nil {
			allFlags = append(allFlags, parsedFlags...)
		}
	}

	if err := util.CreateDir(volumePath); err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create target path %s: %v", volumePath, err)
	}

	klog.V(1).Infof("Mounting bucket %s to %s with: %v", bucket, volumePath, allFlags)
	cmd := exec.Command("gcsfuse", allFlags...)
	cmd.Args = append(cmd.Args, bucket, volumePath)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error (code %s) mounting bucket %s:\n%s", err, bucket, out)
	}

	d.volumes[volumeId] = &GCSVolume{
		bucket:  bucket,
		count:   1,
		flags:   allFlags,
		keyFile: keyFile,
		path:    volumePath,
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (d *GCSDriver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	klog.V(4).Infof("Method NodeUnpublishVolume called with: %+v", req)

	d.m.Lock()
	defer d.m.Unlock()

	volumeId := req.GetVolumeId()
	if volumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	v, exists := d.volumes[volumeId]
	if !exists {
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	v.count--
	if v.count > 0 {
		klog.V(1).Infof("Will not remove volume %s as it is still has %d reference(s)", volumeId, v.count)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}

	cmd := exec.Command("fusermount", "-u", v.path)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error (code %s) un-mounting bucket %s:\n%s", err, v.bucket, out)
	}

	location := filepath.Dir(v.keyFile)
	if location == KeyStoragePath {
		if err := os.Remove(v.keyFile); err != nil {
			klog.Warningf("Error removing temporary key file %s: %s", v.keyFile, err)
		}
	}

	delete(d.volumes, volumeId)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (d *GCSDriver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	klog.V(4).Infof("Method NodeGetInfo called with: %+v", req)

	return &csi.NodeGetInfoResponse{NodeId: d.nodeName}, nil
}

func (d *GCSDriver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	klog.V(4).Infof("Method NodeGetCapabilities called with: %+v", req)

	return &csi.NodeGetCapabilitiesResponse{Capabilities: []*csi.NodeServiceCapability{}}, nil
}

func (d *GCSDriver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "NodeStageVolume: not implemented by %s", d.name)
}

func (d *GCSDriver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "NodeUnstageVolume: not implemented by %s", d.name)
}

func (d *GCSDriver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "NodeGetVolumeStats: not implemented by %s", d.name)
}

func (d *GCSDriver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "NodeExpandVolume: not implemented by %s", d.name)
}
