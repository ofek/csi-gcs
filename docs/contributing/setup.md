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

* Build Docker Image `invoke image`
* Delete Pod

## Build docs

* `invoke docs`


## Sanity Tests

* `invoke test.sanity`
