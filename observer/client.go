package observer

import (
	"github.com/cermakm/argo-await/common"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// AwaitResourceClient watches for specified resources
type AwaitResourceClient struct {
	cl dynamic.NamespaceableResourceInterface

	Log          *logrus.Logger
	K8RestConfig *rest.Config
}

// Namespace to be watched by the client
func (rc *AwaitResourceClient) Namespace(ns string) dynamic.ResourceInterface {
	return rc.cl.Namespace(ns)
}

// NewClientForConfigResource returns a new client based on the passed in kubernetes config
func NewClientForConfigResource(conf *rest.Config, res *metav1.APIResource) (*AwaitResourceClient, error) {
	dynamicClient := dynamic.NewForConfigOrDie(conf)

	gvr := schema.GroupVersionResource{
		Group:    res.Group,
		Version:  res.Version, // apiResource seems to be have empty Version string
		Resource: res.Name,
	}
	resourceClient := dynamicClient.Resource(gvr)

	return &AwaitResourceClient{
		cl:           resourceClient,
		Log:          common.Logger(),
		K8RestConfig: conf,
	}, nil
}
