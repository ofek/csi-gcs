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
kubectl apply -k "github.com/ofek/csi-gcs/examples/static?ref=master"
```

Confirm it's working by running

```console
kubectl get pods,pv,pvc
```

You should see something like

```
NAME                              READY   STATUS    RESTARTS   AGE
pod/csi-gcs-test-cbc546b4-5kb7h   2/2     Running   0          1m40s

NAME                          CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                 STORAGECLASS      REASON   AGE
persistentvolume/csi-gcs-pv   5Gi        RWO            Retain           Bound    default/csi-gcs-pvc   csi-gcs-test-sc            1m40s

NAME                                STATUS   VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS      AGE
persistentvolumeclaim/csi-gcs-pvc   Bound    csi-gcs-pv   5Gi        RWO            csi-gcs-test-sc   1m40s
```

Note the pod name, in this case `csi-gcs-test-cbc546b4-5kb7h`. The pod in the example deployment has 2 containers: a `writer` and a `reader`.

Now create some data!

```console
kubectl exec csi-gcs-test-cbc546b4-5kb7h -c writer -- /bin/sh -c "echo Hello from Google Cloud Storage! > /data/test.txt"
```

Let's read what we just put in the bucket

```
$ kubectl exec csi-gcs-test-cbc546b4-5kb7h -c reader -it -- /bin/sh
/ # ls -lh /data
total 1K
-rw-r--r--    1 root     root          33 Jan 26 17:55 test.txt
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
kubectl delete -k "github.com/ofek/csi-gcs/examples/static?ref=master"
kubectl delete secret csi-gcs-secret
kubectl delete -k "github.com/ofek/csi-gcs/deploy/overlays/stable?ref=master"
```

## Driver options

See the CSI section of the [Kubernetes Volume docs][k8s-volume-csi].

### Service account key

The contents of the JSON key may be passed in as a secret defined in `nodePublishSecretRef`.
The name of the key is `key`

### Bucket

The bucket name is resolved in the following order:

1. `bucket` in `volumeAttributes`
2. `bucket` in secret referenced by `nodePublishSecretRef`
3. `volumeHandle`

### Extra flags

You can pass arbitrary flags to [gcsfuse](https://github.com/GoogleCloudPlatform/gcsfuse) by setting
`flags` in `volumeAttributes` e.g. `--limit-ops-per-sec=10 --only-dir=some/nested/folder`.

## Permission

In order to access anything stored in GCS, you will need [service accounts][gcp-service-account] with
appropriate IAM roles. If writes are needed, you will usually select `roles/storage.objectAdmin` scoped
to the desired buckets.

The [easiest way][gcp-create-service-account] to create service account keys, if you don't yet
have any, is to run:

```console
gcloud iam service-accounts list
```

to find the email of a desired service account, then run:

```console
gcloud iam service-accounts keys create <FILE_NAME>.json --iam-account <EMAIL>
```

to create a key file.

--8<-- "refs.txt"
