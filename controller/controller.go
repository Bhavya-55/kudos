package controller

import (
	"context"
	"fmt"
	"kudos/cloud"
	"kudos/pkg/apis/bhavya.dev/v1alpha"
	"log"

	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	//"k8s.io/apimachinery/pkg/conversion"
	"github.com/kanisterio/kanister/pkg/poll"
)

var cloudResource = schema.GroupVersionResource{
	Group:    "bhavya.dev",
	Version:  "v1alpha",
	Resource: "klusters",
}

type controller struct {
	kclient   kubernetes.Interface
	dclient   dynamic.Interface
	cacheSync cache.InformerSynced
	queue     workqueue.RateLimitingInterface
	kLister   cache.Store
}

func NewController(kclient kubernetes.Interface, dclient dynamic.Interface, dynamicInf dynamicinformer.DynamicSharedInformerFactory) *controller {
	informer := dynamicInf.ForResource(cloudResource).Informer()
	c := &controller{
		kclient:   kclient,
		dclient:   dclient,
		queue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kluster"),
		cacheSync: informer.HasSynced,
		kLister:   informer.GetStore(),
	}
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.add,
		DeleteFunc: c.delete,
	})
	return c
}
func (c *controller) Run(stopCh <-chan struct{}) {
	fmt.Println("starting controller\n")
	if !cache.WaitForCacheSync(stopCh, c.cacheSync) {
		fmt.Println("waiting for cache\n")
	}
	go wait.Until(c.worker, 1*time.Second, stopCh)
	<-stopCh
}
func (c *controller) worker() {
	for c.processNextItem() {
		// this loop helps process item func run until it returns true
	}
}
func (c *controller) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)

	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("error getting key from cache: %v", err)
		return false
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("error splitting key: %v", err)
		return false
	}

	// Try to get the resource
	kluster, err := c.dclient.Resource(cloudResource).Namespace(ns).Get(context.Background(), name, metav1.GetOptions{})

	// If resource is deleted (not found)
	if apierrors.IsNotFound(err) {
		log.Printf("Kluster %s/%s was deleted, cleaning up DO cluster", ns, name)
		// Get the object from item since it's not in API server anymore
		if obj, ok := item.(*unstructured.Unstructured); ok {
			spec := obj.Object["spec"].(map[string]interface{})
			_, err := cloud.DeleteCluster(c.kclient, name, spec)
			if err != nil {
				log.Printf("error deleting DO cluster: %v", err)
				c.queue.AddRateLimited(item)
				return false
			}
			log.Printf("Successfully deleted DO cluster for %s/%s", ns, name)
		}
		return true
	}

	// Handle other errors
	if err != nil {
		log.Printf("error getting kluster: %v", err)
		c.queue.AddRateLimited(item)
		return false
	}

	// Get spec from the resource
	spec, ok := kluster.Object["spec"].(map[string]interface{})
	if !ok {
		log.Printf("error getting spec from kluster")
		return false
	}

	log.Printf("Processing kluster %s/%s", ns, name)

	// Check if this is a new cluster that needs to be created
	if kluster.GetDeletionTimestamp() == nil {
		id, err := cloud.Create(c.kclient, spec)
		if err != nil {
			log.Printf("error creating DO cluster: %v", err)
			c.queue.AddRateLimited(item)
			return false
		}
		log.Printf("Created DO cluster with id %s", id)

		var k v1alpha.Kluster
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(kluster.Object, &k); err != nil {
			log.Printf("error converting unstructured: %v", err)
			return false
		}

		if err := c.updateStatus(id, "creating", &k, k.Status); err != nil {
			log.Printf("error updating status: %v", err)
			return false
		}

		if err := c.WaitForCluster(spec, id); err != nil {
			log.Printf("error waiting for cluster: %v", err)
			return false
		}

		if err := c.updateStatus(id, "running", &k, k.Status); err != nil {
			log.Printf("error updating final status: %v", err)
			return false
		}
	}

	c.queue.Forget(item)
	return true
}
func (c *controller) processNextItem1() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	c.queue.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		log.Printf("error %s calling Namespace key func on cache for item", err.Error())
		return false
	}
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Printf("error %s calling splitNamespace key func on cache for item", err.Error())
		return false
	}
	kluster, err := c.dclient.Resource(cloudResource).Namespace(ns).Get(context.Background(), name, metav1.GetOptions{})
	spec := kluster.Object["spec"].(map[string]interface{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			_, err := cloud.DeleteCluster(c.kclient, name, spec)
			if err != nil {
				log.Printf("error %s,unable to  Delete the kluster resource", err.Error())
				return false
			} else {
				return true
			}
		}
		log.Printf("error %s, Getting the kluster resource from lister", err.Error())
		return false
	}
	//log.Println("kluster spec that we have is %s\n", kluster.spec)

	log.Printf("kluster spec that we have is %s\n", spec)
	id, err := cloud.Create(c.kclient, spec)
	if err != nil {
		log.Printf("error %s, Creating the kluster resource", err.Error())
		return false
	}
	log.Printf("kluster created with id %s\n", id)
	var k v1alpha.Kluster
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(kluster.Object, &k); err != nil {
		log.Printf("error converting unstructured to Kluster: %v", err)
		return false
	}

	if err := c.updateStatus(id, "creating", &k, k.Status); err != nil {
		log.Printf("error updating status: %v", err)
		return false
	}

	if err := c.WaitForCluster(spec, id); err != nil {
		log.Printf("error waiting for cluster: %v", err)
		return false
	}

	if err := c.updateStatus(id, "running", &k, k.Status); err != nil {
		log.Printf("error updating final status: %v", err)
		return false
	}
	return true

}
func (c *controller) WaitForCluster(spec map[string]interface{}, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return poll.Wait(ctx, func(ctx context.Context) (bool, error) {
		state, err := cloud.ClusterStatus(c.kclient, spec, id)
		if err != nil {
			return false, err
		}
		if state == "running" {
			return true, nil
		}
		return false, nil
	})

}
func (c *controller) updateStatus(id, process string, kluster *v1alpha.Kluster, status v1alpha.KlusterStatus) error {
	k, err := c.dclient.Resource(cloudResource).Get(context.Background(), kluster.Name, metav1.GetOptions{})
	if err != nil {
		log.Println("error getting kluster %s", err.Error())
	}
	if k == nil {
		return fmt.Errorf("received nil kluster from dynamic client")
	}
	statusMap := map[string]interface{}{
		"status": map[string]interface{}{
			"process":    status.DeepCopy().Process,
			"klusterId":  status.DeepCopy().KlusterId,
			"kubeConfig": status.DeepCopy().KubeConfig,
		},
	}
	k.Object["status"] = statusMap
	_, err = c.dclient.Resource(cloudResource).UpdateStatus(context.Background(), k, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating status: %w", err)
	}

	return nil
}
func deleteDoCluster() bool {
	fmt.Println(" cluster deleted !")
	return true
}
func (c *controller) add(obj interface{}) {
	fmt.Println("Add event")
	c.queue.Add(obj)
}
func (c *controller) delete(obj interface{}) {
	fmt.Println("delete event")
	c.queue.Add(obj)
}
