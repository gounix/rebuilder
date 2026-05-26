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
	"path/filepath"
	"time"
	"strings"
	"errors"
	//corev1 "k8s.io/api/core/v1"
	//appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
        "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/tools/clientcmd"

        "rebuilder/logger"
        "rebuilder/environ"
	"rebuilder/resources"
)

type RestartPod struct {
	Kind string
	Name string
	StartedAt v1.Time
}

func listPods(clientset *kubernetes.Clientset, namespace string) []RestartPod {
	var list = []RestartPod{}
	var list_entry RestartPod

	podClient := clientset.CoreV1().Pods(namespace)
        podOut, _ := podClient.List(context.TODO(),v1.ListOptions{})
        for _, entry := range podOut.Items {
		owner := entry.ObjectMeta.OwnerReferences
		status := entry.Status.ContainerStatuses
		if len(owner) > 0 && len(status) > 0 {
			list_entry.Kind = owner[0].Kind
			list_entry.Name = owner[0].Name
			if list_entry.Kind == "ReplicaSet" || list_entry.Kind == "Deployment" || list_entry.Kind == "DaemonSet" || list_entry.Kind == "StatefulSet" {
				logger.Info("actions.listPods", "kind", list_entry.Kind, "name", list_entry.Name, "startedAt", status[0].State.Running.StartedAt)
				//fmt.Printf("actions.listPods list_entry.StartedAt %s\n", status)
				list_entry.StartedAt = status[0].State.Running.StartedAt
				list = append(list, list_entry)
			}
		}
        }
        return list
}

func findStatefulSetParent(clientset *kubernetes.Clientset, namespace string, kind string, name string) (string, string, error) {
	client := clientset.AppsV1().StatefulSets(namespace)
	out, _ := client.List(context.TODO(),v1.ListOptions{})

        for _, entry := range out.Items {
		if entry.ObjectMeta.Name == name {
			owner := entry.ObjectMeta.OwnerReferences
			if len(owner) == 0 {
				// return my name
				logger.Info("actions.findStatefulSetParent found myself", "kind", kind, "name", name)
				return kind, name, nil
			} else {
				logger.Info("actions.findStatefulSetParent found", "kind", owner[0].Kind, "name", owner[0].Name)
				return owner[0].Kind, owner[0].Name, nil
			}
		}
	}
	logger.Error("actions.findStatefulSetParent no parent found")
	return "", "", errors.New("actions.findStatefulSetParent no parent found")
}

func findReplicaSetParent(clientset *kubernetes.Clientset, namespace string, kind string, name string) (string, string, error) {
	client := clientset.AppsV1().ReplicaSets(namespace)
	out, _ := client.List(context.TODO(),v1.ListOptions{})

        for _, entry := range out.Items {
		if entry.ObjectMeta.Name == name {
			owner := entry.ObjectMeta.OwnerReferences
			if len(owner) == 0 {
				// return my name
				logger.Info("actions.findReplicaSetParent found myself", "kind", kind, "name", name)
				return kind, name, nil
			} else {
				logger.Info("actions.findReplicaSetParent found", "kind", owner[0].Kind, "name", owner[0].Name)
				return owner[0].Kind, owner[0].Name, nil
			}
		}
	}
	logger.Error("actions.findReplicaSetParent no parent found")
	return "", "", errors.New("actions.findReplicaSetParent no parent found")
}

func findDaemonSetParent(clientset *kubernetes.Clientset, namespace string, kind string, name string) (string, string, error) {
	client := clientset.AppsV1().DaemonSets(namespace)
	out, _ := client.List(context.TODO(),v1.ListOptions{})

        for _, entry := range out.Items {
		if entry.ObjectMeta.Name == name {
			owner := entry.ObjectMeta.OwnerReferences
			if len(owner) == 0 {
				// return my name
				logger.Info("actions.findDaemonSetParent found myself", "kind", kind, "name", name)
				return kind, name, nil
			} else {
				logger.Info("actions.findDaemonSetParent found", "kind", owner[0].Kind, "name", owner[0].Name)
				return owner[0].Kind, owner[0].Name, nil
			}
		}
	}
	logger.Error("actions.findDaemonSetParent no parent found")
	return "", "", errors.New("actions.findDaemonSetParent no parent found")
}

func findParent(clientset *kubernetes.Clientset, namespace string, kind string, name string) (string, string, error) {

	logger.Info("actions.findParent", "kind", kind, "name", name)
	switch kind {
	case "ReplicaSet":
		return findReplicaSetParent(clientset, namespace, kind, name)
	case "StatefulSet":
		return findStatefulSetParent(clientset, namespace, kind, name)
	case "DaemonSet":
		return findDaemonSetParent(clientset, namespace, kind, name)
	default:
		logger.Error("actions.findParent", "unknown kind", kind)
	}
	
	return "", "", errors.New("unknown kind " + kind)
}

func restartAllowed(kind string, name string, actions []resources.ActionsT) bool {

	for _, entry := range actions {
		if strings.EqualFold(kind, entry.Actiontype) && strings.EqualFold(name, entry.Name) {
			logger.Info("actions.restartAllowed true", "kind", kind, "name", name)
			return true
		}
	}
	logger.Info("actions.restartAllowed false", "kind", kind, "name", name)
	return false
}

func RestartNeeded(namespace string, actions []resources.ActionsT, imageTime time.Time) error {
	var kubeconfig string

        if environ.Env.Standalone {
                if home := homedir.HomeDir(); home != "" {
                        kubeconfig = filepath.Join(home, ".kube", "config")
                }
        }

        // use the current context in kubeconfig
        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
                logger.Error("actions.RestartNeeded", "clientcmd.BuildConfigFromFlags", err)
                return err
        }

        // create the clientset
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                logger.Error("actions.RestartNeeded", "kubernetes.NewForConfig", err)
                return err
        }

	// list pods with startdate and owner
	podlist := listPods(clientset, namespace)

	// trace back to deployments, daemonsets, ... 
	for _, entry := range podlist {
		logger.Info("actions.RestartNeeded", "kind", entry.Kind, "name", entry.Name, "startedAt", entry.StartedAt)
		restartKind, restartName, err := findParent(clientset, namespace, entry.Kind, entry.Name)
		if err != nil {
			logger.Error("actions.RestartNeeded", "findParent", err)
			continue
		}
		logger.Info("actions.RestartNeeded parent", "kind", restartKind, "name", restartName)

		entryTime, err := time.Parse("2006-01-02 15:04:05 Z0700 MST", fmt.Sprintf("%s", entry.StartedAt))
		if err != nil {
			logger.Error("actions.RestartNeeded", "time.Parse err", err)
			continue
		}

		// compare with image build date
		// restart the pod if the image was already renewed for another deployment
		if imageTime.After(entryTime) {
			logger.Info("actions.RestartNeeded restarting", "kind", restartKind, "name", restartName, "image time", imageTime, "pod start time", entryTime)
			// if kind and name is present in action list
			if restartAllowed(restartKind, restartName, actions) {
				if err = RunActions(namespace, actions); err != nil {
                                        logger.Error("actions.RestartNeeded", "RunActions error", err)
					continue
                                }

			}
		}
	}
	return nil
}

