package main

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	"flag"
	controller "kudos/controller"

	//"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var cloudResource = schema.GroupVersionResource{
	Group:    "bhavya.dev",
	Version:  "v1alpha",
	Resource: "klusters",
}

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/bhavyaw/.kube/config", "location to kube config file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("error building config: %v\n", err)
		return
	}

	dclient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Printf("error building dynamic client: %v\n", err)
		return
	}

	kclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error building kubernetes client: %v\n", err)
		return
	}

	// Create informer factory with a reasonable resync period
	infFactory := dynamicinformer.NewDynamicSharedInformerFactory(dclient, time.Second*30)

	// Get informer for your resource
	informer := infFactory.ForResource(cloudResource).Informer()

	// Add event handler logging
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Printf("Add event received: %v\n", obj)
		},
		UpdateFunc: func(old, new interface{}) {
			fmt.Printf("Update event received: %v\n", new)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Printf("Delete event received: %v\n", obj)
		},
	})

	c := controller.NewController(kclient, dclient, infFactory)

	// Create stop channel
	stopCh := make(chan struct{})
	defer close(stopCh)

	// Start the informer factory
	infFactory.Start(stopCh)

	// Wait for caches to sync
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		fmt.Println("Failed to sync caches")
		return
	}

	fmt.Println("Cache synced successfully")

	// Run the controller
	c.Run(stopCh)

	// Block forever
	select {}
}
