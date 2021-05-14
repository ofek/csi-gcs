# Troubleshooting

-----

## Warnings events from pods

Warnings, like the one below, can be seen from pods scheduled on newly started nodes.

```
MountVolume.MountDevice failed for volume "xxxx" : kubernetes.io/csi: attacher.MountDevice failed to create newCsiDriverClient: driver name gcs.csi.ofek.dev not found in the list of registered CSI drivers
```

Those warnings are temporary and reflect that the driver is still starting. Kubernetes will retry until the driver is ready.

It's possible to avoid those warnings by adding a node selector or affinity using the node label `<driver name>/driver-ready=true`.
By default `<driver name>` is `gcs.csi.ofek.dev`.

!!! note
    Adding such node selector or affinity will trade the time spend waiting for volume mounting retries with time waiting for scheduling.

```yaml
apiVersion: v1
kind: Pod
metadata:
    name: pod-mount-csi-gcs-volume
spec:
  // ...
  nodeSelector:
    gcs.csi.ofek.dev/driver-ready: "true"
```

```yaml
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

You can also add an admission mutating webhook to automatically inject such node selector or affinity in all pods mounting `csi-gcs` volumes.
