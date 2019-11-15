package main

import (
	"encoding/json"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	//value  是集群文件对应地址
	kubeconfig := flag.String("kubeconfig", "src/config.conf", "")
	flag.Parse()
	//CreateVirtualService(kubeconfig)
	PatchVirtualService(kubeconfig)
}

func virtualServiceInit(kubeconfig *string) (*rest.Config, schema.GroupVersionResource, error) {
	// 解析到config
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, schema.GroupVersionResource{}, err
	}
	//  Create a GVR which represents an Istio Virtual Service.
	virtualServiceGVR := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "virtualservices",
	}
	return config, virtualServiceGVR, nil
}

func virtualServcieApi(kubeconfig *string) {
	config, virtualServiceGVR, err := virtualServiceInit(kubeconfig)
	if err != nil {
		fmt.Println(err.Error())
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
	}
	virtualServices, err := dynamicClient.Resource(virtualServiceGVR).Namespace("default").List(metav1.ListOptions{})
	for index, virtualService := range virtualServices.Items {
		fmt.Printf("VirtualService %d: %s\n", index+1, virtualService.GetName())
	}
}

// Gateway is the generic Kubernetes API Object wrapper
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              map[string]interface{} `json:"spec"`
}

func CreateVirtualService(kubeconfig *string) {
	config, virtualServiceGVR, err := virtualServiceInit(kubeconfig)
	if err != nil {
		fmt.Println(err.Error())
	}
	//tr := &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//}
	//config.TLSClientConfig.Insecure=true
	//config.Transport = tr
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
	}
	var virtualServiceObj unstructured.Unstructured
	//资源分配会遇到无法设置值的问题，故采用json反解析
	j := `{
	"apiVersion": "networking.istio.io/v1alpha3",
    "kind": "VirtualService",
    "metadata": {
        "name": "reviews"
    },
    "spec": {
    	"hosts": ["reviews"],
        "http": [{
        	"route": [{
        		"destination": {
        			"host":"reviews",
        			"subset":"v1"
        		}
        	}]
          }]
    }
}`
	json.Unmarshal([]byte(j), &virtualServiceObj.Object)
	_, err = dynamicClient.Resource(virtualServiceGVR).Namespace("foo").Create(&virtualServiceObj, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("create virtualService success")
}

func UpdateVirtualService(kubeconfig *string) {
	config, virtualServiceGVR, err := virtualServiceInit(kubeconfig)
	if err != nil {
		fmt.Println(err.Error())
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Println(err.Error())
	}
	var virtualServiceObj unstructured.Unstructured
	//资源分配会遇到无法设置值的问题，故采用json反解析
	j := `{
	"apiVersion": "networking.istio.io/v1alpha3",
    "kind": "VirtualService",
    "metadata": {
        "name": "reviews"
    },
    "spec": {
    	"hosts": ["reviews"],
        "http": [{
        	"route": [{
        		"destination": {
        			"host":"reviews",
        			"subset":"v2"
        		}
        	}]
          }]
    }
}`
	json.Unmarshal([]byte(j), &virtualServiceObj.Object)
	virtualServiceName := virtualServiceObj.GetName()
	err = dynamicClient.Resource(virtualServiceGVR).Namespace("foo").Delete(virtualServiceName, &metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err.Error())
	} else {
		_, err = dynamicClient.Resource(virtualServiceGVR).Namespace("foo").Create(&virtualServiceObj, metav1.CreateOptions{})
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("update virtualService success")
		}
	}
}

func DeleteVirtualService(kubeconfig *string) {
	config, virtualServiceGVR, err := virtualServiceInit(kubeconfig)
	if err != nil {
		fmt.Println(err.Error())
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Println(err.Error())
	}
	err = dynamicClient.Resource(virtualServiceGVR).Namespace("foo").Delete("reviews", &metav1.DeleteOptions{})
	if err != nil {
		fmt.Println(err.Error())
	}
}

type PatchUInt32Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func PatchVirtualService(kubeconfig *string) {
	config, virtualServiceGVR, err := virtualServiceInit(kubeconfig)
	if err != nil {
		fmt.Println(err.Error())
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Println(err.Error())
	}
	patchPayload := make([]PatchUInt32Value, 1)
	/*
		spec:
	    hosts:
	    - reviews
	    http:
	    - route:
	      - destination:
	          host: reviews
	          subset: v3
	*/
	patchPayload[0].Op = "replace"
	patchPayload[0].Path = "/spec/http/0/route/0/destination/subset"
	patchPayload[0].Value = "v3"
	patchBytes, _ := json.Marshal(patchPayload)
	_, err = dynamicClient.Resource(virtualServiceGVR).Namespace("foo").Patch("reviews", types.JSONPatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("patch virtualService success")
	}
}
