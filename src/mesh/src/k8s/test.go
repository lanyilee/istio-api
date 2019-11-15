package mygok8s

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"strconv"
	//appsv1beta1 "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/apps/v1beta1"
)

func GetSidecarServiceNum(kubeconfig *string) (allServiceNum int, sidecarServiceNum int) {

	// 解析到config
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// 创建连接
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	//ns all
	var nsList *apiv1.NamespaceList
	nsList, err = clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
	}
	//svc all
	svcList := []apiv1.Service{}
	//pod all
	podList := []apiv1.Pod{}
	if nsList != nil && len(nsList.Items) > 0 {
		for _, ns := range nsList.Items {
			//找到对应的svc
			svcs, err := clientset.CoreV1().Services(ns.Name).List(metav1.ListOptions{})
			if err != nil {
				fmt.Println(err)
			}
			svcList = append(svcList, svcs.Items...)
			//找到对应的pod
			pods, err := clientset.CoreV1().Pods(ns.Name).List(metav1.ListOptions{})
			if err != nil {
				fmt.Println(err)
			}
			podList = append(podList, pods.Items...)
		}
	}
	fmt.Println("Svcs: " + strconv.Itoa(len(svcList)))
	fmt.Println("Pods: " + strconv.Itoa(len(podList)))
	//svc contain sidecar
	svcSidecarList := []apiv1.Service{}
	//pod contain sidecar
	//podSidecarList :=apiv1.Pod{}
	for _, svc := range svcList {
		for key, _ := range svc.Spec.Selector {
			hasAddSidecarList := false
			for _, pod := range podList {
				if svc.Spec.Selector[key] == pod.Labels[key] && svc.Namespace == pod.Namespace {
					for _, container := range pod.Spec.Containers {
						if container.Name == "istio-proxy" {
							svcSidecarList = append(svcSidecarList, svc)
							hasAddSidecarList = true
							break
						}
					}
					break
				}
			}
			if hasAddSidecarList {
				break
			}
		}

	}
	for _, svc := range svcSidecarList {
		fmt.Println(svc.Namespace + ": " + svc.Name)
	}
	fmt.Println(len(svcSidecarList))
	return len(svcList), len(svcSidecarList)
}

//查询service

//监听Deployment变化
func startWatchDeployment(deploymentsClient v1beta1.DeploymentInterface) {
	w, _ := deploymentsClient.Watch(metav1.ListOptions{})
	for {
		select {
		case e, _ := <-w.ResultChan():
			fmt.Println(e.Type, e.Object)
		}
	}
}

func int32Ptr2(i int32) *int32 { return &i }
