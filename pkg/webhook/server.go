package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/ofek/csi-gcs/pkg/util"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

type handler struct {
	k8sClient                   *kubernetes.Clientset
	driverReadyLabel            string
	driverReadySelectorPodPatch []byte
	driverName                  string
}

func NewServer(driverName string) (http.Handler, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	patch := []struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value string `json:"value"`
	}{{
		Op:    "add",
		Path:  "/spec/nodeSelector/" + util.DriverReadyLabelJSONPatchEscaped(driverName),
		Value: "true",
	}}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	h := handler{
		k8sClient:                   clientset,
		driverReadyLabel:            util.DriverReadyLabel(driverName),
		driverReadySelectorPodPatch: patchBytes,
		driverName:                  driverName,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate-inject-driver-ready-selector", h.handleInjectDriverReadySelector)
	mux.HandleFunc("/healthz", h.handleHealthz)

	return mux, nil
}

func (h *handler) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return
}

func (h *handler) handleInjectDriverReadySelector(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	admrev := admissionv1.AdmissionReview{}
	err := json.NewDecoder(r.Body).Decode(&admrev)
	if err != nil {
		http.Error(w, "unable to decode request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if admrev.Request.Operation != admissionv1.Create {
		http.Error(w, "unsupported admission operation, operation must be 'create'", http.StatusBadRequest)
		return
	}

	pod := corev1.Pod{}
	err = json.Unmarshal(admrev.Request.Object.Raw, &pod)
	klog.V(6).Infof("Received '%s'", string(admrev.Request.Object.Raw))
	if err != nil {
		http.Error(w, "unable to decode request object, expected v1/Pod: "+err.Error(), http.StatusBadRequest)
		return
	}
	// The namespace need to be populated from the AdmissionReview Request info
	// as the pod doesn't necessarily has a namespace yet.
	pod.Namespace = admrev.Request.Namespace

	admresp := admissionv1.AdmissionResponse{
		UID:     admrev.Request.UID,
		Allowed: true,
	}
	if podHasDriverReadyLabelSelectorOrAffinity(&pod, h.driverReadyLabel) {
		klog.V(5).Infof("Skipping pod %s/%s already has driver ready preference", pod.Namespace, pod.Name)
	} else {
		if podHasCsiGCSVolume(&pod, h.driverName, h.k8sClient.CoreV1()) {
			klog.V(5).Infof("Mutating pod %s/%s", pod.Namespace, pod.Name)
			patchType := admissionv1.PatchTypeJSONPatch
			admresp.PatchType = &patchType
			admresp.Patch = h.driverReadySelectorPodPatch
		} else {
			klog.V(5).Infof("Skipping pod %s/%s doesn't has csi-gcs volume", pod.Namespace, pod.Name)
		}
	}

	jsonOKResponse(w, &admissionv1.AdmissionReview{
		TypeMeta: admrev.TypeMeta,
		Response: &admresp,
	})
	return
}

func jsonOKResponse(w http.ResponseWriter, rsp interface{}) {
	bts, err := json.Marshal(rsp)
	if err != nil {
		http.Error(w, "unable to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	klog.V(6).Infof("Answering '%s'", string(bts))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bts)
}
