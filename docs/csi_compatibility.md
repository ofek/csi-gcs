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
