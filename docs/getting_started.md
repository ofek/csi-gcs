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
kubectl get CSIDriver,daemonsets,statefulsets,pods -n kube-system
```

should contain something like

```
NAME                                        CREATED AT
csidriver.storage.k8s.io/gcs.csi.ofek.dev   2020-04-19T03:35:52Z

NAME                                DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR                 AGE
daemonset.extensions/csi-gcs-node   1         1         1       1            1           <none>                        4s

NAME                                  READY   AGE
statefulset.apps/csi-gcs-controller   1/1     4s

NAME                                         READY   STATUS    RESTARTS   AGE
pod/csi-gcs-controller-0                     2/2     Running   0          4s
pod/csi-gcs-node-mbmnc                       2/2     Running   0          4s
```

## Debugging

```console
kubectl logs -l app=csi-gcs-controller -c csi-gcs-controller -n kube-system
kubectl logs -l app=csi-gcs-node -c csi-gcs-node -n kube-system
```

## Resource Requests / Limits

To change the default resource requests & limits, override them using kustomize.

**kustomization.yaml**

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
  - github.com/ofek/csi-gcs/deploy/overlays/stable-gke?ref=v0.4.0
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
since the `DaemonSet` requires `priorityClassName` `system-node-critical` to be
prioritized over normal workloads.