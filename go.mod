module github.com/ofek/csi-gcs

go 1.13

require (
	cloud.google.com/go/storage v1.10.0
	github.com/container-storage-interface/spec v1.2.0
	github.com/kubernetes-csi/csi-lib-utils v0.7.0
	github.com/kubernetes-csi/csi-test/v3 v3.1.1-0.20200525083111-e89bc15a6e5e
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	google.golang.org/api v0.43.0
	google.golang.org/grpc v1.36.1
	k8s.io/apimachinery v0.24.0
	k8s.io/client-go v0.24.0
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
)
