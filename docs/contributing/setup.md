# Setup

-----

## Getting started

* Dependencies
    - You'll need to have [Python 3.6+](https://www.python.org/downloads/) in your PATH
    - `python -m pip install --upgrade -r requirements.txt`
* Minikube
    - Setup [`minikube`](https://kubernetes.io/docs/tasks/tools/install-minikube/#installing-minikube)
    - Start `minikube` (`minikube start`)
* Build
    - Enable `minikube` Docker Env (`eval $(minikube docker-env)`)
    - Build Docker Image `invoke image`
* `gcloud`
    - Install [`gcloud`](https://cloud.google.com/sdk/install)
    - Login to `gcloud` (`gcloud auth login`)
* Google Cloud Project
    - Create Test Project (`gcloud projects create [PROJECT_ID] --name=[PROJECT_NAME]`)
* Google Cloud Service Account
    - Create (`gcloud iam service-accounts create [ACCOUNT_NAME] --display-name="Test Account" --description="Test Account for GCS CSI" --project=[PROJECT_ID]`)
    - Create Key (`gcloud iam service-accounts keys create service-account.json --iam-account=[ACCOUNT_NAME]@[PROJECT_ID].iam.gserviceaccount.com  --project=[PROJECT_ID]`)
    - Give Storage Admin Permission (`gcloud projects add-iam-policy-binding [PROJECT_ID] --member=serviceAccount:[ACCOUNT_NAME]@[PROJECT_ID].iam.gserviceaccount.com --role=roles/storage.admin`)
* Create Secret
    - `kubectl create secret generic csi-gcs-secret --from-file=key=service-account.json`
* Pull Needed Images
    - `docker pull quay.io/k8scsi/csi-node-driver-registrar:v1.2.0`
* Apply config `kubectl apply -k deploy/overlays/dev`

## Rebuild & Test Manually in Minikube

```console
# Build Binary
invoke build

# Build Container
invoke image
```

Afterwards kill the currently running pod.

## Documentation

```console
# Build
invoke docs.build

# Server
invoke docs.serve
```


## Sanity Tests

Needs root privileges and `gcsfuse` installed, execution via docker recommended.

```console
# Local
invoke test.sanity

# Docker
invoke docker -c "invoke test.sanity"
```

Additionally the file `./test/secret.yaml` has to be created with the following content:

```yml
CreateVolumeSecret:
  projectId: [Google Cloud Project ID]
  key: |
    [Storage Admin Key JSON]
DeleteVolumeSecret:
  projectId: [Google Cloud Project ID]
  key: |
    [Storage Admin Key JSON]
ControllerPublishVolumeSecret:
  projectId: [Google Cloud Project ID]
  key: |
    [Storage Admin Key JSON]
ControllerUnpublishVolumeSecret:
  projectId: [Google Cloud Project ID]
  key: |
    [Storage Admin Key JSON]
NodeStageVolumeSecret:
  projectId: [Google Cloud Project ID]
  key: |
    [Storage Object Admin Key JSON]
NodePublishVolumeSecret:
  projectId: [Google Cloud Project ID]
  key: |
    [Storage Object Admin Key JSON]
ControllerValidateVolumeCapabilitiesSecret:
  projectId: [Google Cloud Project ID]
  key: |
    [Storage Admin Key JSON]
```

## Develop inside Docker

Run all `invoke` commands through `invoke env -c "[CMD]"`.

## Regenerating the API Client

If any changes are made in the `pkg/apis` package, the `pkg/client` needs to be regenerated.

To regenerate the client package, run `invoke codegen`.
