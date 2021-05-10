package webhook

import (
	"testing"

	"github.com/ofek/csi-gcs/pkg/driver"
	"github.com/ofek/csi-gcs/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	corev1Client "k8s.io/client-go/kubernetes/typed/core/v1"
)

// CSIDriverNameCustom used to test when a non-default driver name is used.
const CSIDriverNameCustom = "custom.driver.name"

func Test_podHasCsiGCSVolume(t *testing.T) {
	t.Parallel()
	type args struct {
		pod        *corev1.Pod
		driverName string
		k8sClient  corev1Client.CoreV1Interface
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Positive tests
		{name: "CSI volume", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "csi-gcs-volume",
							VolumeSource: corev1.VolumeSource{
								CSI: &corev1.CSIVolumeSource{
									Driver: driver.CSIDriverName,
								},
							},
						},
					},
				},
			},
			driver.CSIDriverName,
			nil, // easy way to test that podHasCsiGCSVolume does not make API calls
		}, want: true},
		{name: "CSI volume custom driver name", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "csi-gcs-volume",
							VolumeSource: corev1.VolumeSource{
								CSI: &corev1.CSIVolumeSource{
									Driver: CSIDriverNameCustom,
								},
							},
						},
					},
				},
			},
			CSIDriverNameCustom,
			nil,
		}, want: true},
		{name: "PVC CSI-GCS", args: args{
			&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "csi-gcs-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "csi-gcs-pvc",
								},
							},
						},
					},
				},
			},
			driver.CSIDriverName,
			fake.NewSimpleClientset(&corev1.PersistentVolumeClaim{
				TypeMeta: v1.TypeMeta{
					Kind:       "PersistentVolumeClaim",
					APIVersion: "v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Namespace: "default",
					Name:      "csi-gcs-pvc",
					Annotations: map[string]string{
						"volume.beta.kubernetes.io/storage-provisioner": driver.CSIDriverName,
					},
				},
			}).CoreV1(),
		}, want: true},
		{name: "PVC CSI-GCS with custom driver name", args: args{
			&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "csi-gcs-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "csi-gcs-pvc",
								},
							},
						},
					},
				},
			},
			CSIDriverNameCustom,
			fake.NewSimpleClientset(&corev1.PersistentVolumeClaim{
				TypeMeta: v1.TypeMeta{
					Kind:       "PersistentVolumeClaim",
					APIVersion: "v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Namespace: "default",
					Name:      "csi-gcs-pvc",
					Annotations: map[string]string{
						"volume.beta.kubernetes.io/storage-provisioner": CSIDriverNameCustom,
					},
				},
			}).CoreV1(),
		}, want: true},

		// Negative tests
		{name: "nothing", args: args{
			&corev1.Pod{},
			driver.CSIDriverName,
			nil,
		}, want: false},
		{name: "missing PVC", args: args{
			&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "missing-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "missing-volume",
								},
							},
						},
					},
				},
			},
			driver.CSIDriverName,
			fake.NewSimpleClientset().CoreV1(),
		}, want: false},
		{name: "unrelated volume", args: args{
			&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "unrelated-volume",
							VolumeSource: corev1.VolumeSource{
								GCEPersistentDisk: &corev1.GCEPersistentDiskVolumeSource{
									PDName: "unrelated",
								},
							},
						},
					},
				},
			},
			driver.CSIDriverName,
			nil,
		}, want: false},
		{name: "unrelated PVC", args: args{
			&corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "unrelated-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "unrelated",
								},
							},
						},
					},
				},
			},
			driver.CSIDriverName,
			fake.NewSimpleClientset(&corev1.PersistentVolumeClaim{
				TypeMeta: v1.TypeMeta{
					Kind:       "PersistentVolumeClaim",
					APIVersion: "v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Namespace: "default",
					Name:      "unrelated",
					Annotations: map[string]string{
						"volume.beta.kubernetes.io/storage-provisioner": "gce-pd",
					},
				},
			}).CoreV1(),
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := podHasCsiGCSVolume(tt.args.pod, tt.args.driverName, tt.args.k8sClient); got != tt.want {
				t.Errorf("podHasCsiGCSVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_podHasDriverReadyLabelSelectorOrAffinity(t *testing.T) {
	t.Parallel()
	type args struct {
		pod              *corev1.Pod
		driverReadyLabel string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Positive tests
		{name: "NodeSelector", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						util.DriverReadyLabel(driver.CSIDriverName): "true",
					},
				},
			},
			util.DriverReadyLabel(driver.CSIDriverName),
		}, want: true},
		{name: "NodeSelector not true value", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						util.DriverReadyLabel(driver.CSIDriverName): "definitely not true",
					},
				},
			},
			util.DriverReadyLabel(driver.CSIDriverName),
		}, want: true},
		{name: "NodeSelector and custom driver name", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						CSIDriverNameCustom: "true",
					},
				},
			},
			CSIDriverNameCustom,
		}, want: true},
		{name: "RequiredDuringSchedulingIgnoredDuringExecution", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key: util.DriverReadyLabel(driver.CSIDriverName),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			util.DriverReadyLabel(driver.CSIDriverName),
		}, want: true},
		{name: "RequiredDuringSchedulingIgnoredDuringExecution and custom driver name", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key: CSIDriverNameCustom,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			CSIDriverNameCustom,
		}, want: true},
		{name: "PreferredDuringSchedulingIgnoredDuringExecution", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
								{
									Preference: corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{Key: util.DriverReadyLabel(driver.CSIDriverName)},
										},
									},
								},
							},
						},
					},
				},
			},
			util.DriverReadyLabel(driver.CSIDriverName),
		}, want: true},
		{name: "PreferredDuringSchedulingIgnoredDuringExecution and custom driver name", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
								{
									Preference: corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{Key: CSIDriverNameCustom},
										},
									},
								},
							},
						},
					},
				},
			},
			CSIDriverNameCustom,
		}, want: true},

		// Negative tests
		{name: "PreferredDuringSchedulingIgnoredDuringExecution unrelated", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
								{
									Preference: corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{Key: "unrelated"},
										},
									},
								},
							},
						},
					},
				},
			},
			util.DriverReadyLabel(driver.CSIDriverName),
		}, want: false},
		{name: "RequiredDuringSchedulingIgnoredDuringExecution unrelated", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key: "unrelated",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			util.DriverReadyLabel(driver.CSIDriverName),
		}, want: false},
		{name: "NodeSelector unrelated", args: args{
			&corev1.Pod{
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"unrelated": "unrelated",
					},
				},
			},
			util.DriverReadyLabel(driver.CSIDriverName),
		}, want: false},
		{name: "nothing", args: args{
			&corev1.Pod{},
			util.DriverReadyLabel(driver.CSIDriverName),
		}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := podHasDriverReadyLabelSelectorOrAffinity(tt.args.pod, tt.args.driverReadyLabel); got != tt.want {
				t.Errorf("podHasDriverReadyLabelSelectorOrAffinity() = %v, want %v", got, tt.want)
			}
		})
	}
}
