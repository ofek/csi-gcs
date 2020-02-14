# `csi-gcs`

[![Docker - Pulls](https://img.shields.io/docker/pulls/ofekmeister/csi-gcs.svg)](https://hub.docker.com/r/ofekmeister/csi-gcs)
[![License - MIT/Apache-2.0](https://img.shields.io/badge/license-MIT%2FApache--2.0-9400d3.svg)](https://choosealicense.com/licenses)
[![Say Thanks](https://img.shields.io/badge/say-thanks-ff69b4.svg)](https://saythanks.io/to/ofek)

-----

An easy-to-use, cross-platform, and highly optimized Kubernetes CSI driver for mounting Google Cloud Storage buckets.

**Table of Contents**

- [Getting started](#getting-started)
- [Installation](#installation)
- [Usage](#usage)
- [Debugging](#debugging)
- [Driver options](#driver-options)
  - [Service account key](#service-account-key)
  - [Bucket](#bucket)
  - [Extra flags](#extra-flags)
- [Permission](#permission)
- [Dynamic provisioning](#dynamic-provisioning)
- [License](#license)
- [Disclaimer](#disclaimer)

## Getting started

After acquiring a [service account key](#permission), create a [secret](https://kubernetes.io/docs/concepts/configuration/secret/) (we'll
call it `csi-gcs-secret` in the following example):

```console
kubectl create secret generic csi-gcs-secret --from-literal=bucket=<BUCKET_NAME> --from-file=key=<PATH_TO_SERVICE_ACCOUNT_KEY>
```

Note we store the desired bucket in the secret for brevity only, there are [other ways](#bucket) to select a bucket.

## Installation

Like other CSI drivers, a [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) is the recommended deployment mechanism.

Run

```console
kubectl apply -k github.com/ofek/csi-gcs/deploy/overlays/stable?ref=master
```

Now the output from running the command

```console
kubectl get pods,daemonsets,CSIDriver -n kube-system
```

should contain something like

```
NAME                                         READY   STATUS    RESTARTS   AGE
pod/csi-gcs-95xgx                            2/2     Running   0          1m

NAME                              DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.extensions/csi-gcs      1         1         1       1            1           <none>          1m

NAME                                        CREATED AT
csidriver.storage.k8s.io/gcs.csi.ofek.dev   2020-01-26T15:49:44Z
```

## Usage

Let's run an example application!

```console
kubectl apply -k github.com/ofek/csi-gcs/examples/static?ref=master
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

Notice that while the `writer` container's permission is completely governed by the service account key, the `reader` container is further restricted to read-only access

```
/ # touch /data/forbidden.txt
touch: /data/forbidden.txt: Read-only file system
```

To clean up everything, run the following commands

```console
kubectl delete -k github.com/ofek/csi-gcs/examples/static?ref=master
kubectl delete -k github.com/ofek/csi-gcs/deploy/overlays/stable?ref=master
kubectl delete secret csi-gcs-secret
```

## Debugging

```console
kubectl logs -l app=csi-gcs -c csi-gcs -n kube-system
```

## Driver options

See https://kubernetes.io/docs/concepts/storage/volumes/#csi

### Service account key

The contents of the JSON key may be passed in as a secret defined in `nodePublishSecretRef`. The default secret
name of the key is `key`, though you can select the name by setting `key_name` in `volumeAttributes`.

Alternatively, you can also mount service account keys directly in the driver's container. In this case, set
`key_file` in `volumeAttributes` to the absolute path of the mounted key file.

### Bucket

The bucket name is resolved in the following order:

1. `bucket` in `nodePublishSecretRef`
2. `bucket` in `volumeAttributes`
3. `volumeHandle`

### Extra flags

You can pass arbitrary flags to [gcsfuse](https://github.com/GoogleCloudPlatform/gcsfuse) by setting
`flags` in `volumeAttributes` e.g. `--limit-ops-per-sec=10 --only-dir=some/nested/folder`.

## Permission

In order to access anything stored in Google Cloud Storage, you will need service accounts with appropriate IAM
roles. Read more about them [here](https://cloud.google.com/iam/docs/understanding-service-accounts). If writes
are needed, you will usually select `roles/storage.objectAdmin` scoped to the desired buckets.

The easiest way to create service account keys, if you don't yet have any, is to run:

```console
gcloud iam service-accounts list
```

to find the email of a desired service account, then run:

```console
gcloud iam service-accounts keys create <FILE_NAME>.json --iam-account <EMAIL>
```

to create a key file.

## Dynamic provisioning

Currently, the buckets used must already exist. PRs adding
[this ability](https://cloud.google.com/storage/docs/reference/libraries#client-libraries-usage-go) are extremely welcome!

## License

`csi-gcs` is distributed under the terms of both

- [Apache License, Version 2.0](https://choosealicense.com/licenses/apache-2.0)
- [MIT License](https://choosealicense.com/licenses/mit)

at your option.

## Disclaimer

This is not an official Google product.
