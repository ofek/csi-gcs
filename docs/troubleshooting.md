# Troubleshooting
## Warnings events from pods scheduled on newly started nodes when mounting csi-gcs volumes

Warnings, like the one below, can be seen from pods scheduled on newly started nodes.

```
MountVolume.MountDevice failed for volume "xxxx" : kubernetes.io/csi: attacher.MountDevice failed to create newCsiDriverClient: driver name gcs.csi.ofek.dev not found in the list of registered CSI drivers
```

Those warnings are temporary and reflect the csi-gcs driver is still starting. Kubernetes will retry until the csi-gcs driver is ready.

It's possible to avoid those warnings by adding a node selector or affinity using the node label `gcs.csi.ofek.dev/driver-ready=true`.

> Adding such node selector or affinity will trade the time spend waiting for volume mounting retries against time waiting for scheduling.


```
apiVersion: v1
kind: Pod
metadata:
    name: pod-mount-csi-gcs-volume
spec:
  // ...
  nodeSelector:
    gcs.csi.ofek.dev/driver-ready: "true"
```

```
apiVersion: v1
kind: Pod
metadata:
    name: pod-mount-csi-gcs-volume
spec:
  // ...
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: gcs.csi.ofek.dev/driver-ready
            operator: In
            values:
            - "true"
```

You can also add an admission mutating webhook to automatically inject such node selector or affinity in all pods mounting csi-gcs volumes.
