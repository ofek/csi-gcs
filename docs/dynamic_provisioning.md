# Dynamic provisioning

-----

## Secrets

After acquiring [service account keys](#permission), create 2 [secrets][k8s-secret] (we'll call
them `csi-gcs-secret-mounter` and `csi-gcs-secret-creator` in the following example):

```console
kubectl create secret generic csi-gcs-secret-mounter --from-file=key=<PATH_TO_SERVICE_ACCOUNT_KEY_1>
kubectl create secret generic csi-gcs-secret-creator --from-file=key=<PATH_TO_SERVICE_ACCOUNT_KEY_2> --from-literal=projectId=csi-gcs
```

## Usage

Let's run another example application!

```console
kubectl apply -k "github.com/ofek/csi-gcs/examples/dynamic?ref=master"
```

Confirm it's working by running

```console
kubectl get pods,pv,pvc
```

You should see something like

```
NAME                                READY   STATUS    RESTARTS   AGE
pod/csi-gcs-test-68dbf75685-p7x4g   2/2     Running   0          11s

NAME                                                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                 STORAGECLASS   REASON   AGE
persistentvolume/pvc-3cd15760-b893-40c8-93d4-c93b121c7400   5Gi        RWO            Retain           Bound    default/csi-gcs-pvc   csi-gcs                 10s

NAME                                STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/csi-gcs-pvc   Bound    pvc-3cd15760-b893-40c8-93d4-c93b121c7400   5Gi        RWO            csi-gcs        11s
```

Note the pod name, in this case `csi-gcs-test-68dbf75685-p7x4g`. The pod in the example deployment has 2 containers: a `writer` and a `reader`.

Now create some data!

```console
kubectl exec csi-gcs-test-68dbf75685-p7x4g -c writer -- /bin/sh -c "echo Hello from Google Cloud Storage! > /data/test.txt"
```

Let's read what we just put in the bucket

```
$ kubectl exec csi-gcs-test-68dbf75685-p7x4g -c reader -it -- /bin/sh
/ # ls -lh /data
total 1K
-rw-r--r--    1 root     root          33 Apr 19 16:18 test.txt
/ # cat /data/test.txt
Hello from Google Cloud Storage!
```

Notice that while the `writer` container's permission is completely governed by the `mounter`'s service account key,
the `reader` container is further restricted to read-only access

```
/ # touch /data/forbidden.txt
touch: /data/forbidden.txt: Read-only file system
```

To clean up everything, run the following commands

```console
kubectl delete -f "https://github.com/ofek/csi-gcs/blob/master/examples/dynamic/deployment.yaml"
kubectl delete -f "https://github.com/ofek/csi-gcs/blob/master/examples/dynamic/pvc.yaml"
kubectl delete -f "https://github.com/ofek/csi-gcs/blob/master/examples/dynamic/sc.yaml"
kubectl delete secret csi-gcs-secret-creator
kubectl delete secret csi-gcs-secret-mounter
kubectl delete -k "github.com/ofek/csi-gcs/deploy/overlays/stable?ref=master"
```

??? note
    Cleanup is necessarily verbose until [this](https://github.com/kubernetes-sigs/kustomize/issues/2138) is resolved.

## Driver options

[StorageClass][k8s-storage-class] is the resource type that enables dynamic provisioning.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: <STORAGE_CLASS_NAME>
provisioner: gcs.csi.ofek.dev
reclaimPolicy: Delete
parameters:
  ...
```

### Storage Class Parameters

| Annotation | Description |
| --- | --- |
| `csi.storage.k8s.io/node-publish-secret-name` | The name of the secret allowed to mount created buckets |
| `csi.storage.k8s.io/node-publish-secret-namespace` | The namespace of the secret allowed to mount created buckets |
| `csi.storage.k8s.io/provisioner-secret-name` | The name of the secret allowed to create buckets |
| `csi.storage.k8s.io/provisioner-secret-namespace` | The namespace of the secret allowed to create buckets |
| `csi.storage.k8s.io/controller-expand-secret-name` | The name of the secret allowed to expand [bucket capacity](csi_compatibility.md#capacity) |
| `csi.storage.k8s.io/controller-expand-secret-namespace` | The namespace of the secret allowed to expand [bucket capacity](csi_compatibility.md#capacity) |
| `gcs.csi.ofek.dev/project-id` | The project to create the buckets in. If not specified, `projectId` will be looked up in the provisioner's secret |
| `gcs.csi.ofek.dev/location` | The [location][gcs-location] to create buckets at (default `US` multi-region) |

### Persistent Volume Claim Parameters

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    ...
```

| Annotation | Description |
| --- | --- |
| `gcs.csi.ofek.dev/project-id` | The project to create the buckets in. If not specified, `projectId` will be looked up in the provisioner's secret |
| `gcs.csi.ofek.dev/location` | The [location][gcs-location] to create buckets at (default `US` multi-region) |
| `gcs.csi.ofek.dev/bucket` | The name for the new bucket |

### Persistent buckets

In our example, the dynamically created buckets are deleted during cleanup. If you want the buckets to not be ephemeral,
you can set `reclaimPolicy` to `Retain`.

### Extra flags

You can pass flags to [gcsfuse][gcsfuse-github]. They will be forwarded to [`PersistentVolumeClaim.spec.csi.volumeAttributes`](static_provisioning.md#extra-flags).

The following flags are supported (ordered by precedence):

1. ??? info "**PersistentVolumeClaim.metadata.annotations**"
       ```yaml
       apiVersion: v1
       kind: PersistentVolumeClaim
       metadata:
         annotations:
           gcs.csi.ofek.dev/gid: "63147"
           gcs.csi.ofek.dev/dir-mode: "0775"
           gcs.csi.ofek.dev/file-mode: "0664"
       ```

        | Option | Type | Description |
        | --- | --- | --- |
        | `gcs.csi.ofek.dev/dir-mode` | Octal Integer | Permission bits for directories. (default: 0775) |
        | `gcs.csi.ofek.dev/file-mode` | Octal Integer | Permission bits for files. (default: 0664) |
        | `gcs.csi.ofek.dev/gid` | Integer | GID owner of all inodes. (default: 63147) |
        | `gcs.csi.ofek.dev/uid` | Integer | UID owner of all inodes. (default: -1) |
        | `gcs.csi.ofek.dev/implicit-dirs` | Flag | [Implicitly][gcsfuse-implicit-dirs] define directories based on content. |
        | `gcs.csi.ofek.dev/billing-project` | Text | Project to use for billing when accessing requester pays buckets. |
        | `gcs.csi.ofek.dev/limit-bytes-per-sec` | Integer | Bandwidth limit for reading data, measured over a 30-second window. The default is -1 (no limit). |
        | `gcs.csi.ofek.dev/limit-ops-per-sec` | Integer | Operations per second limit, measured over a 30-second window. The default is 5. Use -1 for no limit. |
        | `gcs.csi.ofek.dev/stat-cache-ttl` | Text | How long to cache StatObject results and inode attributes e.g. `1h`. |
        | `gcs.csi.ofek.dev/type-cache-ttl` | Text | How long to cache name -> file/dir mappings in directory inodes e.g. `1h`. |
        | `gcs.csi.ofek.dev/fuse-mount-options` | Text[] | Additional comma-separated system-specific [mount options][fuse-mount-options]. Be careful! |

1. ??? info "**StorageClass.parameters**"
       ```yaml
       apiVersion: storage.k8s.io/v1
       kind: StorageClass
       parameters:
         gid: "63147"
         dirMode: "0775"
         fileMode: "0664"
       ```

        | Option | Type | Description |
        | --- | --- | --- |
        | `dirMode` | Octal Integer | Permission bits for directories. (default: 0775) |
        | `fileMode` | Octal Integer | Permission bits for files. (default: 0664) |
        | `gid` | Integer | GID owner of all inodes. (default: 63147) |
        | `uid` | Integer | UID owner of all inodes. (default: -1) |
        | `implicitDirs` | Flag | [Implicitly][gcsfuse-implicit-dirs] define directories based on content. |
        | `billingProject` | Text | Project to use for billing when accessing requester pays buckets. |
        | `limitBytesPerSec` | Integer | Bandwidth limit for reading data, measured over a 30-second window. The default is -1 (no limit). |
        | `limitOpsPerSec` | Integer | Operations per second limit, measured over a 30-second window. The default is 5. Use -1 for no limit. |
        | `statCacheTTL` | Text | How long to cache StatObject results and inode attributes e.g. `1h`. |
        | `typeCacheTTL` | Text | How long to cache name -> file/dir mappings in directory inodes e.g. `1h`. |
        | `fuseMountOptions` | Text[] | Additional comma-separated system-specific [mount options][fuse-mount-options]. Be careful! |

1. ??? info "**StorageClass.mountOptions**"
       ```yaml
       apiVersion: storage.k8s.io/v1
       kind: StorageClass
       mountOptions:
        - --gid=63147
        - --dir-mode=0775
        - --file-mode=0664
       ```

        | Option | Type | Description |
        | --- | --- | --- |
        | `dir-mode` | Octal Integer | Permission bits for directories. (default: 0775) |
        | `file-mode` | Octal Integer | Permission bits for files. (default: 0664) |
        | `gid` | Integer | GID owner of all inodes. (default: 63147) |
        | `uid` | Integer | UID owner of all inodes. (default: -1) |
        | `implicit-dirs` | Flag | [Implicitly][gcsfuse-implicit-dirs] define directories based on content. |
        | `billing-project` | Text | Project to use for billing when accessing requester pays buckets. |
        | `limit-bytes-per-sec` | Integer | Bandwidth limit for reading data, measured over a 30-second window. The default is -1 (no limit). |
        | `limit-ops-per-sec` | Integer | Operations per second limit, measured over a 30-second window. The default is 5. Use -1 for no limit. |
        | `stat-cache-ttl` | Text | How long to cache StatObject results and inode attributes e.g. `1h`. |
        | `type-cache-ttl` | Text | How long to cache name -> file/dir mappings in directory inodes e.g. `1h`. |
        | `fuse-mount-option` | Text | Additional system-specific [mount option][fuse-mount-options]. Be careful! |

1. ??? info "**StorageClass.parameters."csi.storage.k8s.io/provisioner-secret-name**""
       | Option | Type | Description |
       | --- | --- | --- |
       | `dirMode` | Octal Integer | Permission bits for directories, in octal. (default: 0775) |
       | `fileMode` | Octal Integer | Permission bits for files, in octal. (default: 0664) |
       | `gid` | Integer | GID owner of all inodes. (default: 63147) |
       | `uid` | Integer | UID owner of all inodes. (default: -1) |
       | `implicitDirs` | Flag | [Implicitly][gcsfuse-implicit-dirs] define directories based on content. |
       | `billingProject` | Text | Project to use for billing when accessing requester pays buckets. |
       | `limitBytesPerSec` | Integer | Bandwidth limit for reading data, measured over a 30-second window. The default is -1 (no limit). |
       | `limitOpsPerSec` | Integer | Operations per second limit, measured over a 30-second window. The default is 5. Use -1 for no limit. |
       | `statCacheTTL` | Text | How long to cache StatObject results and inode attributes e.g. `1h`. |
       | `typeCacheTTL` | Text | How long to cache name -> file/dir mappings in directory inodes e.g. `1h`. |
       | `fuseMountOptions` | Text[] | Additional comma-separated system-specific [mount options][fuse-mount-options]. Be careful! |

## Permission

In order to access anything stored in GCS, you will need [service accounts][gcp-service-account] with
appropriate IAM roles. You will usually assign the role `roles/storage.admin`.

The [easiest way][gcp-create-sa-key] to create service account keys, if you don't yet
have any, is to run:

```console
gcloud iam service-accounts list
```

to find the email of a desired service account, then run:

```console
gcloud iam service-accounts keys create <FILE_NAME>.json --iam-account <EMAIL>
```

to create a key file.

### Mounter

The [Node Plugin][csi-deploy-node] is the component that is actually mounting and serving buckets to pods.
If writes are needed, you will usually select `roles/storage.objectAdmin` scoped to the desired buckets.

### Creator

The [Controller Plugin][csi-deploy-controller] is the component that is in charge of creating buckets.
The service account will need the `storage.buckets.create` [Cloud IAM permission][gcs-iam-permission].
