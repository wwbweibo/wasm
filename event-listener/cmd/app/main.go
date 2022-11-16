package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"k8s.io/client-go/kubernetes"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/google/uuid"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	dClient, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer dClient.Close()

	config, err := clientcmd.BuildConfigFromFlags("", "/Users/weibo/.kube/config")
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	registerPubSub()
	handlePubSub(clientset)
	_ = http.ListenAndServe(":8080", nil)
}

func registerPubSub() {
	http.HandleFunc("/dapr/subscribe", func(writer http.ResponseWriter, request *http.Request) {
		sub := `
[
		{
			"pubsubname": "pubsub",
			"topic": "test-topic",
			"route": "test-topic"
		}
]
		`
		writer.Write([]byte(sub))
	})
}

func handlePubSub(client *kubernetes.Clientset) {
	http.HandleFunc("/test-topic", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Printf("got message from dapr\n")
		eventBody := map[string]interface{}{}
		err := json.NewDecoder(request.Body).Decode(&eventBody)
		fmt.Printf("event is %s\n", eventBody["data"])
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		data, _ := json.Marshal(eventBody["data"])
		job := batchv1.Job{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Job",
				APIVersion: "batch/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-topic-job-" + uuid.New().String(),
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-topic-job",
								Image: "ghcr.io/wwbweibo/event-handler:latest",
								Args:  []string{string(data)},
							},
						},
						RestartPolicy: corev1.RestartPolicyNever,
						ImagePullSecrets: []corev1.LocalObjectReference{
							{
								Name: "pull-secret",
							},
						},
						Tolerations: []corev1.Toleration{
							{
								Key:      "kubernetes.io/arch",
								Operator: corev1.TolerationOpEqual,
								Value:    "wasm32-wasi",
								Effect:   corev1.TaintEffectNoExecute,
							},
							{
								Key:      "kubernetes.io/arch",
								Operator: corev1.TolerationOpEqual,
								Value:    "wasm32-wasi",
								Effect:   corev1.TaintEffectNoSchedule,
							},
						},
					},
				},
			},
		}

		_, err = client.BatchV1().Jobs("default").Create(context.TODO(), &job, metav1.CreateOptions{})

		if err != nil {
			fmt.Printf("%s\n", err)
		}
	})
}
