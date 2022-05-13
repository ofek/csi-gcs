module github.com/ofek/csi-gcs

go 1.13

require (
	cloud.google.com/go/storage v1.6.0
	github.com/container-storage-interface/spec v1.5.0
	github.com/kubernetes-csi/csi-lib-utils v0.11.0
	github.com/kubernetes-csi/csi-test/v3 v3.1.1-0.20200525083111-e89bc15a6e5e
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.20.0
	google.golang.org/grpc v1.38.0
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v0.22.0
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20210707171843-4b05e18ac7d9
)
