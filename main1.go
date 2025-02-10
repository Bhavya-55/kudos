package main

import (
	"fmt"

	"k8s.io/client-go/kubernetes"

	"flag"
	controller "kudos/controller"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// var cloudResource = schema.GroupVersionResource{
// 	Group:    "bhavya.dev",
// 	Version:  "v1alpha",
// 	Resource: "klusters",
// }

func mainn() {
	kubeconfig := flag.String("kubeconfig", "/home/bhavyaw/.kube/config", "location to kube config file")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Println("error building config %s", err)
	}
	dclient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Println("error building dynamic client %s", err)
	}
	kclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("error building kubernetes client %s", err)
	}
	infFactory := dynamicinformer.NewDynamicSharedInformerFactory(dclient, 0)
	informer := infFactory.ForResource(cloudResource).Informer()

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
	stopCh := make(chan struct{})
	defer close(stopCh)
	infFactory.Start(make(<-chan struct{}))
	c.Run(make(<-chan struct{}))
	//dyInformer.Start(stopCh)
	//stopCh := make(chan struct{})
	// dyInformer.Start(stopCh)
	// Add event handler logging
	// Get informer for your resource

	fmt.Printf("length of kluster we have is %s\n", len(infFactory.ForResource(cloudResource).Informer().GetStore().List()))
	// if err := c.Run(stopCh); err != nil {
	// 	log.Printf("error running controller %s\n", err.Error())
	// }
}
