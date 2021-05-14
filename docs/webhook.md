# Mutating webhook

When scheduling pods using a csi-gcs volume on nodes where csi-gcs drivers aren't ready yet, pods will fail to start until the driver is ready.
This generates warnings events that:
- Creates noise for Kubernetes administrators.
- Can miss-lead some Kubernetes users.

To minimize the problem, the mutating webhook inject the node selector `<driver name>/driver-ready="true"` when needed.  
Preventing pods to schedule on nodes where csi-gcs drivers aren't ready.  

!!! info
    The default `<driver name>` is `gcs.csi.ofek.dev`.

## When is the node selector injected?
The node selector is only injected if
- There is no node affinity using the `<driver name>/driver-ready` label
- The is no node selector using the `<driver name>/driver-ready` label
- There is a csi-gcs volume detected

## What constitutes a pod with a csi-gcs volume?
- A pod with a CSI volume `.spec.volumes[*].csi.driver="<driver name>"`
- A pod with an PersistentVolumeClaim `.spec.volumes[*].PersistentVolumeClaim` annotated with `volume.beta.kubernetes.io/storage-provisioner=<driver name>`

## What are the trade-offs?
- The added latency by the mutation should in the order of milliseconds.
- The added latency waiting for nodes to be ready is equal to or less than the time spent waiting for the pod mount retries.

## How to install the mutating webhook?
TODO