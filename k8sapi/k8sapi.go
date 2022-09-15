package k8sapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"webapp/types"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
)

var deploymentsClient v1.DeploymentInterface
var deploymentJson string

func init() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// TODO: we need to update Istio's virtualservice and destinationrule too

	deploymentsClient = clientset.AppsV1().Deployments("ryogokpoc")

	deploymentJson = `{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "<ToBeReplaced>",
    "labels": {
      "app": "compute",
      "partition": "<ToBeReplaced>"
    }
  },
  "spec": {
    "replicas": 1,
    "selector": {
      "matchLabels": {
        "app": "compute",
        "partition": "<ToBeReplaced>"
      }
    },
    "template": {
      "metadata": {
        "labels": {
          "app": "compute",
          "partition": "<ToBeReplaced>"
        }
      },
      "spec": {
        "serviceAccountName": "pocservice-compute",
        "containers": [
          {
            "name": "compute",
            "image": "ryogokacr.azurecr.io/compute:v1.0",
            "imagePullPolicy": "IfNotPresent",
            "env": [
              {
                "name": "POD_NAME",
                "valueFrom": {
                  "fieldRef": {
                    "fieldPath": "metadata.name"
                  }
                }
              },
              {
                "name": "POD_IP",
                "valueFrom": {
                  "fieldRef": {
                    "fieldPath": "status.podIP"
                  }
                }
              }
            ],
            "ports": [
              {
                "containerPort": 8081
              }
            ]
          }
        ]
      }
    }
  }
}`
}

func UpdateDeployment(modelName string, pinfo *types.PartitionInfo) error {
	log.Println("UpdateDeployment() called")

	// Parse JSON template string into the internal k8s structs
	dec := json.NewDecoder(strings.NewReader(deploymentJson))
	var deployment appsv1.Deployment
	dec.Decode(&deployment)

	// TODO: handle the case where the existing partition was updated
	// For now, do nothing if the partition is not new
	if !pinfo.IsNew {
		return nil
	}

	// Update deployment with real partition information
	deployment.ObjectMeta.Name = "compute-" + pinfo.Name
	deployment.ObjectMeta.Labels["partition"] = pinfo.Name
	deployment.Spec.Selector.MatchLabels["partition"] = pinfo.Name
	deployment.Spec.Template.ObjectMeta.Labels["partition"] = pinfo.Name

	_, err := deploymentsClient.Create(context.TODO(), &deployment, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
