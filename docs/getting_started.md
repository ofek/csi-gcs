# Getting started

-----

## Installation

Like other CSI drivers, a [StatefulSet][k8s-statefulset] and [DaemonSet][k8s-daemonset] are the recommended
deployment mechanisms for the [Controller Plugin][csi-deploy-controller] and [Node Plugin][csi-deploy-node],
respectively.

Run

```console
kubectl apply -k "github.com/ofek/csi-gcs/deploy/overlays/stable?ref=master"
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
