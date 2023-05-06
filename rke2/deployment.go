package rke2

import (
	"context"
	"log"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

// Applies a deployment object to a k8s cluster
func ApplyDeployment(clientset *kubernetes.Clientset, depObj *appsv1.Deployment) error {

	depClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	log.Println("Creating deployment")

	result, err := depClient.Create(context.TODO(), depObj, metav1.CreateOptions{})
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	return nil
}
