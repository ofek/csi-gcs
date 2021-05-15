# Troubleshooting

-----

## Warnings events from pods

Warnings, like the one below, can be seen from pods scheduled on newly started nodes.

```
MountVolume.MountDevice failed for volume "xxxx" : kubernetes.io/csi: attacher.MountDevice failed to create newCsiDriverClient: driver name gcs.csi.ofek.dev not found in the list of registered CSI drivers
```

Those warnings are temporary and reflect that the driver is still starting. Kubernetes will retry until the driver is ready. The problem is often encountered in clusters with auto-scaler as nodes come and go.

This is a known issue of kubernetes (see [kubernetes/issues/75890](https://github.com/kubernetes/kubernetes/issues/75890)).

A possible workaround is to taint all nodes running the csi-gcs driver like `<driver name>/driver-ready=false:NoSchedule` and use, as suggested in [kubernetes/issues/75890#issuecomment-725792993](https://github.com/kubernetes/kubernetes/issues/75890#issuecomment-725792993), a custom controller like [wish/nodetaint](https://github.com/wish/nodetaint) to remove the taint once the csi-gcs pod is ready.

This workaround will ensure pods are repelled from nodes until the csi-gcs driver is ready without interfering with other components like the cluster auto-scaler.

!!! warning
    csi-gcs label the node with `<driver name>/driver-ready=true` to reflect its readiness state. It's possible to use a node selector to select nodes with a ready csi-gcs node driver. However, it doesn't work with clusters using cluster-autoscaler as the auto-scaler will never find a node with matching `<driver name>/driver-ready=true` label. 
