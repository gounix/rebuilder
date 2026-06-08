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

package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/rest"
	"os"
	"path/filepath"
	"rebuilder/environ"
        "rebuilder/logger"
)

var ClientSet *kubernetes.Clientset

func InitConfig() {
        var config *rest.Config
        var err error
        var kubeconfig string

        if environ.Env.Standalone {
                if home := homedir.HomeDir(); home != "" {
                        kubeconfig = filepath.Join(home, ".kube", "config")
                } else {
			logger.Error("k8s.InitConfig could not find kubeconfig")
			os.Exit(1)
                }

                config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
                if err != nil {
			logger.Error("k8s.InitConfig", "error loading given file", err)
			os.Exit(1)
                }
        } else {
                config, err = rest.InClusterConfig()
                if err != nil {
			logger.Error("k8s.InitConfig", "rest.InClusterConfig", err)
			os.Exit(1)
                }
        }

        ClientSet, err = kubernetes.NewForConfig(config)
        if err != nil {
                logger.Error("k8s.InitConfig", "kubernetes.NewForConfig", err)
                os.Exit(1)
        }
}

