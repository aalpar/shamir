// Package scheme provides a Kubernetes runtime scheme with shamir CRDs registered.
package scheme

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	thresholdv1alpha1 "github.com/aalpar/shamir/api/v1alpha1"
)

// Scheme is the runtime scheme with all required types registered.
var Scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))
	utilruntime.Must(thresholdv1alpha1.AddToScheme(Scheme))
}
