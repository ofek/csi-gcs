module github.com/ofek/csi-gcs

go 1.13

require (
	cloud.google.com/go v0.38.0
	github.com/container-storage-interface/spec v1.2.0
	github.com/kubernetes-csi/csi-lib-utils v0.7.0
	github.com/kubernetes-csi/csi-test/v3 v3.1.1-0.20200525083111-e89bc15a6e5e
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.4.0
	google.golang.org/grpc v1.46.0
	k8s.io/apimachinery v0.17.1-beta.0
	k8s.io/client-go v0.17.0
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20191114184206-e782cd3c129f
)
