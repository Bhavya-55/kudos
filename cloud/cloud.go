package cloud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func Create(client kubernetes.Interface, spec map[string]interface{}) (string, error) {
	// Get the token from the secret
	token, err := getToken(client, spec["tokensecret"].(string))
	if token == "" {
		return "", errors.New("error getting token")
	}

	// Initialize the DigitalOcean client
	godoClient := godo.NewFromToken(token)

	// Extract the node pool details
	nodePools, ok := spec["nodepools"].([]interface{})
	if !ok {
		return "", errors.New("error getting nodepools array")
	}

	if len(nodePools) == 0 {
		return "", errors.New("no nodepools specified")
	}

	// Get the first nodepool
	nodePool, ok := nodePools[0].(map[string]interface{})
	if !ok {
		return "", errors.New("error getting nodepool details")
	}

	// Extract count as int64
	count, ok := nodePool["count"].(int64)
	if !ok {
		return "", errors.New("error getting nodepool count")
	}

	// Create the cluster request
	createRequest := &godo.KubernetesClusterCreateRequest{
		Name:        spec["name"].(string),
		RegionSlug:  spec["region"].(string),
		VersionSlug: spec["version"].(string),
		NodePools: []*godo.KubernetesNodePoolCreateRequest{
			{
				Name:  nodePool["name"].(string),
				Size:  nodePool["size"].(string),
				Count: int(count),
			},
		},
	}

	// Create the cluster
	cluster, _, err := godoClient.Kubernetes.Create(context.Background(), createRequest)
	if err != nil {
		return "", err
	}

	return cluster.ID, nil
}
func ClusterStatus(client kubernetes.Interface, spec map[string]interface{}, id string) (string, error) {
	token, err := getToken(client, spec["tokensecret"].(string))
	if token == "" {
		return "", errors.New("error getting token")
	}
	godoClient := godo.NewFromToken(token)
	cluster, _, err := godoClient.Kubernetes.Get(context.Background(), id)
	if err != nil {
		return "", err
	}
	return string(cluster.Status.State), nil
}
func getToken(client kubernetes.Interface, sec string) (string, error) {
	namespace := strings.Split(sec, "/")[0]
	name := strings.Split(sec, "/")[1]
	secret, err := client.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("error getting secret %s", err)
		return "", err

	}
	return string(secret.Data["token"]), nil
}
func retriveCluster(client kubernetes.Interface, token string) ([]*godo.KubernetesCluster, error) {

	godoclient := godo.NewFromToken(token)
	ctx := context.TODO()

	opt := &godo.ListOptions{
		Page:    1,
		PerPage: 200,
	}

	clusters, _, err := godoclient.Kubernetes.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	return clusters, nil
}
func retriveClusterId(client kubernetes.Interface, token string, name string) (string, error) {
	clusters, err := retriveCluster(client, token)
	if err != nil {
		return "", fmt.Errorf("error getting clusters: %v", err)
	}

	for _, c := range clusters {
		if c.Name == name {
			return c.ID, nil
		}
	}

	return "", fmt.Errorf("cluster with name %s not found", name)
}

func DeleteCluster(client kubernetes.Interface, name string, spec map[string]interface{}) (string, error) {
	token, err := getToken(client, spec["tokensecret"].(string))
	if err != nil {
		return "", fmt.Errorf("error getting token: %v", err)
	}
	godoclient := godo.NewFromToken(token)
	ctx := context.TODO()

	clusterId, err := retriveClusterId(client, token, name)
	if err != nil {
		return "", err
	}

	if clusterId == "" {
		return "", fmt.Errorf("invalid cluster ID")
	}

	_, err = godoclient.Kubernetes.Delete(ctx, clusterId)
	if err != nil {
		log.Printf("error deleting cluster: %v", err)
		return "", err
	}

	return "Cluster deleted successfully", nil
}

// func DeleteCluster(client kubernetes.Interface, name string) (string, error) {
// 	token := os.Getenv("DIGITAL_OCEAN_TOKEN")
// 	godoclient := godo.NewFromToken(token)
// 	ctx := context.TODO()
// 	clusterId, err := retriveClusterId(client, token, name)
// 	if err != nil {
// 		return "", err
// 	}
// 	_, err = godoclient.Kubernetes.Delete(ctx, clusterId)
// 	if err != nil {
// 		log.Printf("error deleting cluster %s", err.Error())
// 		return "", err
// 	}
// 	return "Cluster deleted successfully", nil

// }
// func retriveClusterId(client kubernetes.Interface, token string, name string) (string, error) {
// 	clusters, err := retriveCluster(client, token)
// 	if err != nil {
// 		errors.New("error getting cluster %s" + err.Error())

// 	}
// 	for _, c := range clusters {
// 		if c.Name == name {
// 			fmt.Println(true)
// 			return c.ID, nil
// 		}

// 	}
// 	return "", fmt.Errorf("cluster with name %s not found", name)
// }

//--------------------------------------------------------------------------------------------

func Create1(client kubernetes.Interface, spec map[string]interface{}) (string, error) {
	// Get the token from the secret
	token, err := getToken(client, spec["tokensecret"].(string))
	if token == "" {
		return "", errors.New("error getting token")
	}

	// Initialize the DigitalOcean client
	godoClient := godo.NewFromToken(token)
	nodePools := spec["nodePools"].([]interface{})
	nodePool := nodePools[0].(map[string]interface{})
	ct := nodePool["count"].(string)
	count, _ := strconv.Atoi(ct)
	// // Extract the node pool details
	// nodePool, ok := spec["nodepools"].(map[string]interface{})
	// if !ok {
	// 	return "", errors.New("error getting node pool")
	// }
	// ct := nodePool["count"].(string)
	// count, _ := strconv.Atoi(ct)

	// Create the cluster request
	createRequest := &godo.KubernetesClusterCreateRequest{
		Name:        spec["name"].(string),
		RegionSlug:  spec["region"].(string),
		VersionSlug: spec["version"].(string),
		NodePools: []*godo.KubernetesNodePoolCreateRequest{
			{
				Name:  nodePool["name"].(string),
				Size:  nodePool["size"].(string),
				Count: count,
			},
		},
	}

	// Create the cluster
	cluster, _, err := godoClient.Kubernetes.Create(context.Background(), createRequest)
	if err != nil {
		return "", err
	}

	return cluster.ID, nil
}
