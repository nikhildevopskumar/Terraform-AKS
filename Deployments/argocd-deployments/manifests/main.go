package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Create the Nginx pod
	nginxPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-pod",
			Labels: map[string]string{
				"app": "nginx",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx:latest",
				},
			},
		},
	}

	// Create the Busybox pod
	busyboxPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "busybox-pod",
			Labels: map[string]string{
				"app": "busybox",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox:latest",
					Command: []string{"sh", "-c", "while true; do echo Hello from Busybox; sleep 3600; done"},
				},
			},
		},
	}

	// Create a service for the Nginx pod
	nginxService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-service",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "nginx",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}

	// Create a service for the Busybox pod
	busyboxService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "busybox-service",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "busybox",
			},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}

	// Create a NetworkPolicy to allow traffic between the Nginx and Busybox pods
	networkPolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "allow-nginx-busybox-communication",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "busybox",
								},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Port: &intstr.IntOrString{IntVal: 80},
						},
					},
				},
			},
		},
	}

	// Create the pods
	podsClient := clientset.CoreV1().Pods("default")
	_, err = podsClient.Create(context.TODO(), nginxPod, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	_, err = podsClient.Create(context.TODO(), busyboxPod, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Create the services
	servicesClient := clientset.CoreV1().Services("default")
	_, err = servicesClient.Create(context.TODO(), nginxService, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	_, err = servicesClient.Create(context.TODO(), busyboxService, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Create the network policy
	networkPoliciesClient := clientset.NetworkingV1().NetworkPolicies("default")
	_, err = networkPoliciesClient.Create(context.TODO(), networkPolicy, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Created Nginx and Busybox pods, services, and network policy.")
}
