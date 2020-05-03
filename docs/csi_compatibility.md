# CSI Specification Compatibility

This page describes compatibility to the [CSI specification](https://github.com/container-storage-interface/spec/blob/master/spec.md).

## Capacity

!!! warning "Important"
    Google Cloud Storage has no concept of capacity limits. Therefore, this driver is unable to provide capacity limit enforcement.

The driver only sets a `capacity` label for the `bucket` containing the requested bytes.

## Snapshots

[Snapshots](https://github.com/container-storage-interface/spec/blob/master/spec.md#createsnapshot) are not currently supported, but are on the roadmap for the future.

## `CreateVolume` / `VolumeContentSource`

[`CreateVolume` / `VolumeContentSource`](https://github.com/container-storage-interface/spec/blob/master/spec.md#createvolume) is not currently supported, but is on the roadmap for the future.

## Fuse

Since [`gcsfuse`][gcsfuse-github] is backed by [`fuse`][libfuse-github], the mount needs a process to back it. This is an unsolved problem with CSI. See https://github.com/kubernetes/kubernetes/issues/70013

Because of this problem, all mounts will terminate if a pod of the `csi-gcs-node` DaemonSet is restarted. This for example happens when the driver is updated.

To counteract the problem of having pods with broken mounts, the `csi-gcs-node` Pod will terminate all Pods with broken mounts on start.

??? info "Disabling Pod Termination"

    The Pod Termination can be disabled by changing the argument `delete-orphaned-pods` to `false` on the DaemonSet.