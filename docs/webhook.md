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
For development see the [setup guide](contributing/setup.md), in production, you can:
 - Use [cert-manager](https://cert-manager.io/docs/) and inject the TLS certificate automatically. 
 - Generate a self-signed certificate in the same way as for development.

### With cert-manager (recommended for production)
[cert-manager](https://cert-manager.io/docs/) help you managing TLS cert, it can mint and inject TLS certificate automatically.

- [cert-manager installation](https://cert-manager.io/docs/installation/)
- [cert-manager ca injection](https://cert-manager.io/docs/concepts/ca-injector/)

The essential steps are:

Creating a self-signed issuer, if you are already a user of cert-manager it probably already exist 
```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
```

Creating a certificate 
```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: csi-gcs-webhook 
spec:
  dnsNames:
  - <csi-gcs-webhook service name>.<csi-gcs-webhook namesapce>.svc
  issuerRef:
    kind: Issuer
    name: selfsigned-issuer
  secretName: csi-gcs-webhook
```

Patching the `MutatingWebhookConfiguration` to make cert-manager inject the CA automatically. 
```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: gcs.csi.ofek.dev
  annotations:
    cert-manager.io/inject-ca-from: <certificate namesapce>/<certificate name>
```

Patching the webhook server deployment to load the TLS cert and key.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: csi-gcs-webhook-server
spec:
  template:
    spec:
      volumes:
        - name: cert
          secret:
            defaultMode: 420
            secretName: csi-gcs-webhook
      containers:
        - name: server
          volumeMounts:
              - mountPath: /tls
                name: cert
                readOnly: true
```


### With a self-signed certificate

!!! important
The generated certificate is valid for 365 days, you must generate a new one and redeploy the webhook before it expires.

1. Dependencies
    - You'll need to have [Python 3.6+](https://www.python.org/downloads/) in your PATH
    - `python -m pip install --upgrade -r requirements.txt`
2. Generate the tls certificate
    - `invoke tls-cert --namespace=<namespace where the webhook server run> --service=<name of the webhook server service> --output=/some/path`
3. Patch your manifest

You can see how to patch your manifest with kustomize in the [dev overlays](https://github.com/ofek/csi-gcs/tree/master/deploy/overlays/dev).

The essentials steps are:

Creating a `kubernetes.io/tls` secret with the TLS cert and key
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: csi-gcs-webhook
type: kubernetes.io/tls
data:
  tls.crt: <TLS CERT>
  tls.key: <TLS KEY>
```

Patching the mutating webhook `clientConfig.caBundle` to set the TLS cert.
```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: gcs.csi.ofek.dev
webhooks:
  - name: inject-driver-ready-selector.gcs.csi.ofek.dev
    clientConfig:
      caBundle: <TLS CERT>
```

Patching the webhook server deployment to load the TLS cert and key.
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: csi-gcs-webhook-server
spec:
  template:
    spec:
      volumes:
        - name: cert
          secret:
            defaultMode: 420
            secretName: csi-gcs-webhook
      containers:
        - name: server
          volumeMounts:
              - mountPath: /tls
                name: cert
                readOnly: true
```
