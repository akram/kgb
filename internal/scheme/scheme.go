package scheme

import (
	kgbv1alpha1 "github.com/akram/kgb/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// New returns a runtime.Scheme with core Kubernetes + KGB APIs registered.
func New() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(s)
	_ = kgbv1alpha1.AddToScheme(s)
	return s
}
