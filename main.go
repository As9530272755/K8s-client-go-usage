package main

import (
	"context"
	"fmt"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// K8S 配置初始化
func k8sConfig() *kubernetes.Clientset {
	kubeconfig := "etc/config"

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return clientSet
}

// namespace 获取
func GetNamespace() []string {
	namespaceClient := k8sConfig().CoreV1().Namespaces()
	namespaceList, err := namespaceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	namesapces := []string{}
	for _, namespace := range namespaceList.Items {
		namesapces = append(namesapces, namespace.Name)
	}
	return namesapces

}

// deployment 获取
func GetDeployment() {
	for _, namespaces := range GetNamespace() {
		deploymentClient := k8sConfig().AppsV1().Deployments(namespaces)
		deploymentList, err := deploymentClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		} else {
			for _, deployment := range deploymentList.Items {
				fmt.Printf("namespace:%s name:%s\n", deployment.Namespace, deployment.Name)
			}
		}
	}
}

// 由于在创建 deployment 的时候需要指定副本集，而且是一个 int32 的指针类型
func Int32ptr(i int32) *int32 {
	return &i
}

// 创建 deployment
func CreateDeployment() {
	// 通过 config 生成 K8S Deployments.client
	deploymentClient := k8sConfig().AppsV1().Deployments("web")

	// 对 deployment 的操作是 create
	deployment, err := deploymentClient.Create(context.TODO(), &appsv1.Deployment{
		// meta 字段，定义这个 deployment 的 name 和所在的 namespace 以及标签
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-golang",
			Namespace: "web",
			Labels:    map[string]string{"app": "test-golang"},
		},

		// spec 字段，定义副本数等
		Spec: appsv1.DeploymentSpec{
			Replicas: Int32ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test-golang"},
			},
			// template 字段
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test-golang"},
				},

				// template.spec 字段
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "test",
							Image:           "nginx:1.16.1",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}, metav1.CreateOptions{})

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("deployment Create success!\n%v\n", &deployment.Status)
	}

}

// deployment 删除
func DelDeployment() {
	deploymentClient := k8sConfig().AppsV1().Deployments("default")
	err := deploymentClient.Delete(context.TODO(), "golang", metav1.DeleteOptions{})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("web deployment delete success!")
	}
}

// deployment 修改副本集
func EditDeployment() {
	deploymentClient := k8sConfig().AppsV1().Deployments("web")
	deployment, err := deploymentClient.Get(context.TODO(), "test-golang", metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	if *deployment.Spec.Replicas > 2 {
		deployment.Spec.Replicas = Int32ptr(*deployment.Spec.Replicas - 1)
	} else {
		deployment.Spec.Replicas = Int32ptr(*deployment.Spec.Replicas + 1)
	}

	deployment, err = deploymentClient.Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("deployment : %s ,replocas=%d update success!\n", deployment.Name, *deployment.Spec.Replicas)
	}

}

// deployment 修改容器镜像
func EditImage() {
	deploymentClient := k8sConfig().AppsV1().Deployments("web")
	deployment, err := deploymentClient.Get(context.TODO(), "test-golang", metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	deployment.Spec.Template.Spec.Containers[0].Image = "nginx:1.18.0"

	deployment, err = deploymentClient.Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("deployment : %s image update success!\n", deployment.Name)
	}

}

// svc 获取
func GetSvc() {
	for _, namespace := range GetNamespace() {
		svcClient := k8sConfig().CoreV1().Services(namespace)
		svcList, err := svcClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		} else {
			// 获取 svc 基础信息
			for _, svc := range svcList.Items {
				fmt.Printf("\n-------------\nNameSpace:%s\nSvcName:%s\nClusterIP:%s\nLabels:%s\n",
					svc.Namespace, svc.Name, svc.Spec.ClusterIP, svc.Labels)
				// 获取 svc port 信息
				for _, port := range svc.Spec.Ports {
					fmt.Printf("Protocol:%s\nPodPort:%d\nNodePort:%d\n", port.Protocol, port.Port, port.NodePort)
				}
			}
		}
	}
}

// svc 创建
func CreateSvc() {
	svcClient := k8sConfig().CoreV1().Services("web")
	svc, err := svcClient.Create(context.TODO(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "go-nginx-svc",
			Namespace: "web",
			Labels:    map[string]string{"svc": "go-nginx"},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "test-golang"},
			Type:     corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     80,
					Protocol: corev1.ProtocolTCP,
					NodePort: 6110,
				},
			},
		},
	},
		metav1.CreateOptions{})

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("svc: %s create success!\n", svc.Name)
	}
}

func main() {
	CreateSvc()
}
