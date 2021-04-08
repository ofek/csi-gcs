# Static provisioning

-----

## Secrets

After acquiring a [service account key](#permission), create a [secret][k8s-secret] (we'll call
it `csi-gcs-secret` in the following example):

```console
kubectl create secret generic csi-gcs-secret --from-literal=bucket=<BUCKET_NAME> --from-file=key=<PATH_TO_SERVICE_ACCOUNT_KEY>
```

Note we store the desired bucket in the secret for brevity only, there are [other ways](#bucket) to select a bucket.

## Usage

Let's run an example application!

```console
kubectl apply -k "github.com/ofek/csi-gcs/examples/static?ref=<STABLE_VERSION>"
```

Confirm it's working by running

```console
kubectl get pods,pv,pvc
```

You should see something like

```
NAME                                READY   STATUS    RESTARTS   AGE
pod/csi-gcs-test-5f677df9f9-f59km   2/2     Running   0          10s

NAME                          CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                 STORAGECLASS      REASON   AGE
persistentvolume/csi-gcs-pv   5Gi        RWO            Retain           Bound    default/csi-gcs-pvc   csi-gcs-test-sc            10s

NAME                                STATUS   VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS      AGE
persistentvolumeclaim/csi-gcs-pvc   Bound    csi-gcs-pv   5Gi        RWO            csi-gcs-test-sc   10s
```

Note the pod name, in this case `csi-gcs-test-5f677df9f9-f59km`. The pod in the example deployment has 2 containers: a `writer` and a `reader`.

Now create some data!

```console
kubectl exec csi-gcs-test-5f677df9f9-f59km -c writer -- /bin/sh -c "echo Hello from Google Cloud Storage! > /data/test.txt"
```

Let's read what we just put in the bucket

```
$ kubectl exec csi-gcs-test-5f677df9f9-f59km -c reader -it -- /bin/sh
/ # ls -lh /data
total 1K
-rw-r--r--    1 root     root          33 May 26 21:23 test.txt
/ # cat /data/test.txt
Hello from Google Cloud Storage!
```

Notice that while the `writer` container's permission is completely governed by the service account key,
the `reader` container is further restricted to read-only access

```
/ # touch /data/forbidden.txt
touch: /data/forbidden.txt: Read-only file system
```

To clean up everything, run the following commands

```console
kubectl delete -k "github.com/ofek/csi-gcs/examples/static?ref=<STABLE_VERSION>"
kubectl delete -k "github.com/ofek/csi-gcs/deploy/overlays/stable?ref=<STABLE_VERSION>"
kubectl delete secret csi-gcs-secret
```

## Driver options

See the CSI section of the [Kubernetes Volume docs][k8s-volume-csi].

### Service account key

The contents of the JSON key may be passed in as a secret defined in
`PersistentVolume.spec.csi.nodePublishSecretRef`. The name of the key in the secret is `key`.

> You also can just delete the secret definition, the code will automatically fine the service account key follow some order.
> Refer: https://pkg.go.dev/golang.org/x/oauth2/google#FindDefaultCredentials

### Bucket

The bucket name is resolved in the following order:

1. `bucket` in `PersistentVolume.spec.csi.volumeAttributes`
1. `bucket` in `PersistentVolume.spec.mountOptions`
1. `bucket` in secret referenced by `PersistentVolume.spec.csi.nodePublishSecretRef`
1. `PersistentVolume.spec.csi.volumeHandle`

### Extra flags

You can pass flags to [gcsfuse][gcsfuse-github] in the following ways (ordered by precedence):

1. ??? info "**PersistentVolume.spec.csi.volumeAttributes**"
       ```yaml
       apiVersion: v1
       kind: PersistentVolume
       spec:
         csi:
           driver: gcs.csi.ofek.dev
           volumeAttributes:
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

1. ??? info "**PersistentVolume.spec.mountOptions**"
       ```yaml
       apiVersion: v1
       kind: PersistentVolume
       spec:
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
        | `fuse-mount-option` | Text | Additional comma-separated system-specific [mount option][fuse-mount-options]. Be careful! |

1. ??? info "**PersistentVolume.spec.csi.nodePublishSecretRef**"
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
appropriate IAM roles. If writes are needed, you will usually select `roles/storage.objectAdmin` scoped
to the desired buckets.

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
