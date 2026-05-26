/*
MIT License

Copyright (c) 2026 gounix

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package actions

import (
	"fmt"
	"context"
	"time"
	"errors"
	//"encoding/json"
	"path/filepath"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
        "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"rebuilder/logger"
	"rebuilder/resources"
	"rebuilder/environ"
)

const restartAnnotationPattern = `{
			"spec": {
				"template": {
					"metadata": {
						"annotations": {
							"kubectl.kubernetes.io/restartedAt": "%s"
						}
					}
				}
			}
		}`


func restartDeployment(clientset *kubernetes.Clientset, namespace string, name string) error {

	client := clientset.AppsV1().Deployments(namespace)
	data := fmt.Sprintf(restartAnnotationPattern, time.Now().Format(time.RFC3339))
	logger.Info("actions.restartDeployment restarting", "namespace", namespace, "name", name)
	_, err := client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(data), v1.PatchOptions{})

	return err
}

func restartStatefulSet(clientset *kubernetes.Clientset, namespace string, name string) error {

	client := clientset.AppsV1().StatefulSets(namespace)
	data := fmt.Sprintf(restartAnnotationPattern, time.Now().Format(time.RFC3339))
	logger.Info("actions.restartStatefulSet restarting", "namespace", namespace, "name", name)
	_, err := client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(data), v1.PatchOptions{})

	return err
}

func restartDaemonSet(clientset *kubernetes.Clientset, namespace string, name string) error {

	client := clientset.AppsV1().DaemonSets(namespace)
	data := fmt.Sprintf(restartAnnotationPattern, time.Now().Format(time.RFC3339))
	logger.Info("actions.restartDaemonSet restarting", "namespace", namespace, "name", name)
	_, err := client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(data), v1.PatchOptions{})

	return err
}

func restartReplicaSet(clientset *kubernetes.Clientset, namespace string, name string) error {

	client := clientset.AppsV1().ReplicaSets(namespace)
	data := fmt.Sprintf(restartAnnotationPattern, time.Now().Format(time.RFC3339))
	logger.Info("actions.restartReplicaSet restarting", "namespace", namespace, "name", name)
	_, err := client.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(data), v1.PatchOptions{})

	return err
}

func runSimpleAction(clientset *kubernetes.Clientset, namespace string, action resources.ActionsT) error {
	switch action.Actiontype {
	case "deployment":
		return restartDeployment(clientset, namespace, action.Name)
	case "statefulset":
		return restartStatefulSet(clientset, namespace, action.Name)
	case "daemonset":
		return restartDaemonSet(clientset, namespace, action.Name)
	case "replicaset":
		return restartDaemonSet(clientset, namespace, action.Name)
	default:
		logger.Error("actions.runSimpleAction", "action not supported", action.Actiontype)
		return errors.New("action not supported")
	}
	return nil
}

func RunActions(namespace string, actions []resources.ActionsT) error {
	var kubeconfig string

	if environ.Env.Standalone {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

        // use the current context in kubeconfig
        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
                logger.Error("actions.RunActions", "clientcmd.BuildConfigFromFlags", err)
                return err
        }

        // create the clientset
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                logger.Error("actions.RunActions", "kubernetes.NewForConfig", err)
                return err
        }

	for _, entry := range actions {
		logger.Info("actions.RunActions", "namespace", namespace, "type", entry.Actiontype, "name", entry.Name)
		if err := runSimpleAction(clientset, namespace, entry); err != nil {
			logger.Info("actions.RunActions", "namespace", namespace, "type", entry.Actiontype, "name", entry.Name, "err", err)
			return err
		}
	}
	return nil
}

