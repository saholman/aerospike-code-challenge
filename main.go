package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const namespaceName = "aerospike"

type SimplePod struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func SimplePodFromPod(pod v1.Pod) SimplePod {
	return SimplePod{
		Name:      pod.ObjectMeta.Name,
		Namespace: pod.ObjectMeta.Namespace,
	}
}

func main() {
	ctx := context.Background()

	// Connect to the K8s cluster
	log.Info().Msg("connecting to K8s")
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to build config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create clientset from config")
	}

	err = Run(ctx, clientset)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run Aerospike Code Challenge")
	}

	err = Cleanup(ctx, clientset)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to cleanup after Aerospike Code Challenge")
	}
}

func Run(ctx context.Context, clientset kubernetes.Interface) error {
	// Setup client-go informer for Pod events
	informerFactory := informers.NewSharedInformerFactory(clientset, 0)
	podInformer := informerFactory.Core().V1().Pods()
	informer := podInformer.Informer()
	go informerFactory.Start(ctx.Done())
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onPodAdd,
		UpdateFunc: onPodUpdate,
		DeleteFunc: onPodDelete,
	})

	// Print out all namespaces
	nsClient := clientset.CoreV1().Namespaces()
	namespaceList, err := nsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %w", err)
	}
	namespaces := make([]string, 0, len(namespaceList.Items))
	for _, ns := range namespaceList.Items {
		namespaces = append(namespaces, ns.Name)
	}
	log.Info().Interface("namespaces", namespaces).Msg("k8s namespaces")

	// Create a new namespace
	nsSpec := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: namespaceName},
	}
	_, err = nsClient.Create(ctx, nsSpec, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create %s namespace: %w", namespaceName, err)
	}
	log.Info().Msgf("created new namespace %s", namespaceName)

	// Create a pod in that namespace that runs simple hello-world container
	podSpec := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello-world",
			Namespace: namespaceName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name:  "hello-world",
					Image: "hello-world",
				},
			},
		},
	}

	podClient := clientset.CoreV1().Pods(namespaceName)
	_, err = podClient.Create(ctx, podSpec, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}
	log.Info().Msg("created hello-world pod")

	// Print out pod names and the namespaces they are in for any pods that have given label
	label := "k8s-app=kube-dns"
	podList, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		return fmt.Errorf("failed to list pods with label %s: %w", label, err)
	}
	pods := make([]SimplePod, 0, len(podList.Items))
	for _, pod := range podList.Items {
		pods = append(pods, SimplePodFromPod(pod))
	}
	log.Info().Interface("pods", pods).Msgf("pods with label %s", label)
	return nil
}

func Cleanup(ctx context.Context, clientset kubernetes.Interface) error {
	// Delete the hello-world pod created from above
	err := clientset.CoreV1().Pods(namespaceName).Delete(ctx, "hello-world", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}
	log.Info().Msg("deleted hello-world pod")

	// Delete namespace
	err = clientset.CoreV1().Namespaces().Delete(ctx, namespaceName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}
	log.Info().Msgf("deleted %s namespace", namespaceName)
	return nil
}

func onPodAdd(pod any) {
	log.Info().Interface("pod", SimplePodFromPod(*pod.(*v1.Pod))).Msg("informer received pod creation event")
}

func onPodUpdate(oldPod any, newPod any) {
	log.Info().
		Interface("old", SimplePodFromPod(*oldPod.(*v1.Pod))).
		Interface("new", SimplePodFromPod(*newPod.(*v1.Pod))).
		Msg("informer received pod updated event")
}

func onPodDelete(pod any) {
	log.Info().Interface("pod", SimplePodFromPod(*pod.(*v1.Pod))).Msg("informer received pod deletion event")
}
