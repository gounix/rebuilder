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

package resources

import (
	"os"
	"context"
	"fmt"
	"path/filepath"
	"encoding/json"
	"rebuilder/logger"
	"rebuilder/environ"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	api = "gounix.nl"
	api_version = "v1"
	kind = "rebuilds"
)

type (
	MetadataT struct {
		Uid string `json:"uid"`
		Name string `json:"name"`
		Namespace string `json:"namespace"`
	}
	BaseT struct {
		Host string `json:"host"`
		Type string `json:"type"`
		Image string `json:"image"`
		Tag string `json:"tag"`
	}
	GitT struct {
		Host string `json:"host"`
		Project string `json:"project"`
		User string `json:"user"`
		Dir string `json:"dir"`
		SecretName string `json:"secretName"`
		SshKeyName string `json:"sshKeyName"`
	}
	RegistryT struct {
		Host string `json:"host"`
		Type string `json:"type"`
		Image string `json:"image"`
		Tag string `json:"tag"`
		Authenticated bool `json:"authenticated"`
		SecretName string `json:"secretName"`
	}
	ActionsT struct {
		Actiontype string `json:"objecttype"`
		Name string `json:"name"`
	}
	SpecT struct {
		Base BaseT `json:"base"`
		Git GitT `json:"git"`
		Registry RegistryT `json:"registry"`
		Actions []ActionsT `json:"actions"`
	}
	RebuildT struct {
		Metadata MetadataT `json:"metadata"`
		Spec SpecT `json:"spec"`
	}
	RebuildListT struct {
		ApiVersion string `json:"apiVersion"`
		Items []RebuildT  `json:"items"`
	}
)

func GetList() RebuildListT {
	// get the list of rebuild resources from k8s
	var dat RebuildListT
	var kubeconfig string

	if environ.Env.Standalone {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
			// if kubeconfig is empty, BuildConfigFromFlags uses the incluster config
		}
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logger.Error("resources.GetList", "clientcmd.BuildConfigFromFlags", err)
		os.Exit(1)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("resources.GetList", "kubernetes.NewForConfig", err)
		os.Exit(1)
	}

	url := fmt.Sprintf("/apis/%s/%s/%s/", api, api_version, kind)
	out, err := clientset.RESTClient().Get().AbsPath(url).DoRaw(context.TODO())
	if err != nil {
		logger.Error("resources.GetList", "clientset.RESTClient", err)
		os.Exit(1)
	}

	err = json.Unmarshal(out, &dat)
        if err != nil {
		logger.Error("resources.GetList", "unmarshal error", err)
		//os.Exit(1)
	}

	return dat
}
