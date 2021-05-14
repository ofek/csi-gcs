package util

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/ofek/csi-gcs/pkg/apis/published-volume/v1beta1"
	gcs "github.com/ofek/csi-gcs/pkg/client/clientset/clientset"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

func ParseEndpoint(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("could not parse endpoint: %v", err)
	}

	var address string
	if len(u.Host) == 0 {
		address = filepath.FromSlash(u.Path)
	} else {
		address = path.Join(u.Host, filepath.FromSlash(u.Path))
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme == "unix" {
		if err := os.Remove(address); err != nil && !os.IsNotExist(err) {
			return "", "", fmt.Errorf("could not remove unix socket %q: %v", address, err)
		}
	} else {
		return "", "", fmt.Errorf("unsupported protocol: %s", scheme)
	}

	return scheme, address, nil
}

func CreateFile(path, contents string) (string, error) {
	tmpFile, err := ioutil.TempFile(path, "")
	if err != nil {
		return "", fmt.Errorf("error creating file: %s", err)
	}

	filePath := tmpFile.Name()
	fileContents := []byte(contents)

	if _, err := tmpFile.Write(fileContents); err != nil {
		return "", fmt.Errorf("error writing to file %s: %s", filePath, err)
	}

	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("error closing file %s: %s", filePath, err)
	}

	return filePath, nil
}

func CreateDir(d string) error {
	stat, err := os.Lstat(d)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(d, os.ModePerm); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if stat != nil && !stat.IsDir() {
		return fmt.Errorf("%s already exists and is not a directory", d)
	}

	return nil
}

func GetKey(secrets map[string]string, keyStoragePath string) (string, error) {
	if _, err := os.Stat(keyStoragePath); os.IsNotExist(err) {
		os.Mkdir(keyStoragePath, 0700)
	}

	keyContents, keyNameExists := secrets["key"]
	if !keyNameExists {
		return "", status.Errorf(codes.Internal, "Secret '%s' is unavailable", "key")
	}

	klog.V(5).Info("Saving key contents to a temporary location")
	keyFile, err := CreateFile(keyStoragePath, keyContents)
	if err != nil {
		return "", status.Errorf(codes.Internal, "Unable to save secret 'key' to %s", keyStoragePath)
	}

	return keyFile, nil
}

func CleanupKey(keyFile string, keyStoragePath string) {
	location := filepath.Dir(keyFile)
	if location == keyStoragePath {
		if err := os.Remove(keyFile); err != nil {
			klog.Warningf("Error removing temporary key file %s: %s", keyFile, err)
		}
	}
}

func BucketName(volumeId string) string {
	// return volumeId
	var crc32Hash = crc32.ChecksumIEEE([]byte(volumeId))

	if len(volumeId) > 48 {
		volumeId = volumeId[0:48]
	}
	return fmt.Sprintf("%s-%x", strings.ToLower(volumeId), crc32Hash)
}

func BucketCapacity(attrs *storage.BucketAttrs) (int64, error) {
	for labelName, labelValue := range attrs.Labels {
		if labelName != "capacity" {
			continue
		}

		capacity, err := strconv.ParseInt(labelValue, 10, 64)
		if err != nil {
			return 0, status.Errorf(codes.Internal, "Failed to parse bucket capacity: %v", labelValue)
		}

		return capacity, nil
	}

	return 0, nil
}

func SetBucketCapacity(ctx context.Context, bucket *storage.BucketHandle, capacity int64) (attrs *storage.BucketAttrs, err error) {
	var uattrs = storage.BucketAttrsToUpdate{}

	uattrs.SetLabel("capacity", strconv.FormatInt(capacity, 10))

	return bucket.Update(ctx, uattrs)
}

func BucketExists(ctx context.Context, bucket *storage.BucketHandle) (exists bool, err error) {
	query := &storage.Query{Prefix: ""}

	it := bucket.Objects(ctx, query)
	_, err = it.Next()

	if err == iterator.Done {
		return true, nil
	} else if err != nil && err.Error() == "storage: bucket doesn't exist" {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func GetPvcAnnotations(pvcName string, pvcNamespace string) (annotations map[string]string, err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	pvc, err := clientset.CoreV1().PersistentVolumeClaims(pvcNamespace).Get(pvcName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pvc.ObjectMeta.Annotations, nil
}

// DriverReadyLabel returns the driver-ready label according to the driver name.
func DriverReadyLabel(driverName string) string {
	return driverName + "/driver-ready"
}

// DriverReadyLabelJSONPatchEscaped returns the driver-ready label according to the driver name but espcaed to be used in a JSONPatch path.
func DriverReadyLabelJSONPatchEscaped(driverName string) string {
	return strings.ReplaceAll(DriverReadyLabel(driverName), "/", "~1")
}

// SetDriverReadyLabel set the label <driver name>/driver-ready=<isReady> on the given node.
func SetDriverReadyLabel(driverName string, nodeName string, isReady bool) (err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	patch := []struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value string `json:"value"`
	}{{
		Op:    "add",
		Path:  "/metadata/labels/" + DriverReadyLabelJSONPatchEscaped(driverName),
		Value: strconv.FormatBool(isReady),
	}}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	_, err = clientset.CoreV1().Nodes().Patch(nodeName, types.JSONPatchType, patchBytes)
	if err != nil {
		return err
	}
	return nil
}

func DeletePod(namespace string, name string) (err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return clientset.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{})
}

func GetRegisteredMounts(node string) (list *v1beta1.PublishedVolumeList, err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// creates the clientset
	clientset, err := gcs.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset.GcsV1beta1().PublishedVolumes().List(metav1.ListOptions{
		LabelSelector: labels.Set(map[string]string{
			"gcs.csi.ofek.dev/node": node,
		}).String(),
	})
}

func RegisterMount(volumeID string, targetPath string, node string, podNamespace string, podName string, options map[string]string) (err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	// creates the clientsets
	clientset, err := gcs.NewForConfig(config)
	if err != nil {
		return err
	}
	coreClientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	name := strconv.FormatUint(uint64(crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s-%s-%s", volumeID, targetPath, node)))), 16)

	nodeResource, err := coreClientset.CoreV1().Nodes().Get(node, metav1.GetOptions{})
	if err != nil {
		return err
	}

	_, err = clientset.GcsV1beta1().PublishedVolumes().Create(&v1beta1.PublishedVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"gcs.csi.ofek.dev/node": node,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "v1",
					Kind:       "Node",
					Name:       node,
					UID:        nodeResource.GetUID(),
				},
			},
		},
		Spec: v1beta1.PublishedVolumeSpec{
			Node:         node,
			TargetPath:   targetPath,
			VolumeHandle: volumeID,
			Options:      options,
			Pod: v1beta1.PublishedVolumeSpecPod{
				Namespace: podNamespace,
				Name:      podName,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func UnregisterMount(volumeID string, targetPath string, node string) (err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	// creates the clientset
	clientset, err := gcs.NewForConfig(config)
	if err != nil {
		return err
	}

	name := strconv.FormatUint(uint64(crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s-%s-%s", volumeID, targetPath, node)))), 16)

	delPropPolicy := metav1.DeletePropagationForeground
	err = clientset.GcsV1beta1().PublishedVolumes().Delete(name, &metav1.DeleteOptions{
		PropagationPolicy: &delPropPolicy,
	})
	if err != nil {
		return err
	}

	return nil
}

func SetEnvVarFlags() {
	flagset := flag.CommandLine

	// I wish Golang had sets
	set := map[string]string{}

	// https://golang.org/pkg/flag/#FlagSet.Visit
	flagset.Visit(func(f *flag.Flag) {
		set[f.Name] = ""
	})

	// https://golang.org/pkg/flag/#FlagSet.VisitAll
	flagset.VisitAll(func(f *flag.Flag) {
		envVar := strings.Replace(strings.ToUpper(f.Name), "-", "_", -1)

		if val := os.Getenv(envVar); val != "" {
			if _, defined := set[f.Name]; !defined {
				_ = flagset.Set(f.Name, val)
			}
		}

		// Display it in the help text too
		f.Usage = fmt.Sprintf("%s [%s]", f.Usage, envVar)
	})
}
