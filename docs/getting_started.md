# Getting started

-----

## Installation

Like other CSI drivers, a [StatefulSet][k8s-statefulset] and [DaemonSet][k8s-daemonset] are the recommended
deployment mechanisms for the [Controller Plugin][csi-deploy-controller] and [Node Plugin][csi-deploy-node],
respectively.

Run

```console
kubectl apply -k "github.com/ofek/csi-gcs/deploy/overlays/stable?ref=<STABLE_VERSION>"
```

Now the output from running the command

```console
kubectl get CSIDriver,daemonsets,pods -n kube-system
```

should contain something like

```
NAME                                        CREATED AT
csidriver.storage.k8s.io/gcs.csi.ofek.dev   2020-05-26T21:03:14Z

NAME                        DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR                 AGE
daemonset.apps/csi-gcs      1         1         1       1            1           kubernetes.io/os=linux        18s

NAME                                         READY   STATUS    RESTARTS   AGE
pod/csi-gcs-f9vgd                            4/4     Running   0          18s
```

## Customer-managed encryption keys (CMEK)

Make sure that your Google Cloud Storage service account has `roles/cloudkms.cryptoKeyEncrypterDecrypter` for the target encryption key.

`kmsKeyId`/`gcs.csi.ofek.dev/kms-key-id` could be defined as part of a secret or annotation/mount to enable [CMEK encryption for Google Storage](https://cloud.google.com/storage/docs/gsutil/addlhelp/UsingEncryptionKeys).



## Debugging

```console
kubectl logs -l app=csi-gcs -c csi-gcs -n kube-system
```

## Resource Requests / Limits

To change the default resource requests & limits, override them using kustomize.

**kustomization.yaml**

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
  - github.com/ofek/csi-gcs/deploy/overlays/stable-gke?ref=<STABLE_VERSION>
patchesStrategicMerge:
  - resources.yaml
```

**resources.yaml**

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: csi-gcs
spec:
  template:
    spec:
      containers:
      - name: csi-gcs
        resources:
          limits:
            cpu: 1
            memory: 1Gi
          requests:
            cpu: 10m
            memory: 80Mi
```

## Namespace

This driver deploys directly into the `kube-system` namespace. That can't be changed
since the `DaemonSet` requires `priorityClassName: system-node-critical` to be
prioritized over normal workloads.
