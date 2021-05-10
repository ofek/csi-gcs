package webhook

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1Client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
)

// podHasDriverReadyLabelSelectorOrAffinity return true if the pod have any kind of label selector or affinity related to driver readiness.
// It checks:
// 	- node selector
// 	- required node affinity
//  - preferred node affinity
func podHasDriverReadyLabelSelectorOrAffinity(pod *corev1.Pod, driverReadyLabel string) bool {
	if _, exist := pod.Spec.NodeSelector[driverReadyLabel]; exist {
		return true
	}

	if pod.Spec.Affinity != nil &&
		pod.Spec.Affinity.NodeAffinity != nil {

		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			for _, term := range pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
				for _, exp := range term.MatchExpressions {
					if exp.Key == driverReadyLabel {
						return true
					}
				}
			}
		}

		if pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
			for _, term := range pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				for _, exp := range term.Preference.MatchExpressions {
					if exp.Key == driverReadyLabel {
						return true
					}
				}
			}
		}
	}
	return false
}

// podHasCsiGCSVolume return true if a Pod use a csi-gcs volume.
func podHasCsiGCSVolume(pod *corev1.Pod, driverName string, k8sClientCore corev1Client.CoreV1Interface) bool {
	// The checks of the presence of gcs.csi.ofek.dev volume are done in two time
	// The first check if there isn't a PV directly mounted through the CSI spec.
	// It requires no extra network calls.
	for idx, _ := range pod.Spec.Volumes {
		if pod.Spec.Volumes[idx].CSI != nil && pod.Spec.Volumes[idx].CSI.Driver == driverName {
			return true
		}
	}
	// then if there isn't a PVC that belong the csi-gcs.
	// This check require a one network call per PVC.
	for idx, _ := range pod.Spec.Volumes {
		if pod.Spec.Volumes[idx].PersistentVolumeClaim == nil {
			continue
		}
		// TODO when upgrading to k8s api +0.18.0, pass the context to Get()
		pvc, err := k8sClientCore.PersistentVolumeClaims(pod.Namespace).Get(
			pod.Spec.Volumes[idx].PersistentVolumeClaim.ClaimName,
			// if optimization is required, metav1.GetOptions{ResourceVersion: "0"} could probably be set here,
			// as a PVC driver usually don't change much
			metav1.GetOptions{},
		)
		if err != nil {
			klog.Warningf("Unable to fetch PersistentVolumeClaim '%s/%s': %v", pod.Namespace, pod.Spec.Volumes[idx].PersistentVolumeClaim.ClaimName, err)
			continue
		}
		provisioner, exist := pvc.Annotations["volume.beta.kubernetes.io/storage-provisioner"]
		if !exist || provisioner != driverName {
			continue
		}
		return true
	}

	return false
}
