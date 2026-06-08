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

package main

import (
	"fmt"
	"rebuilder/actions"
	"rebuilder/environ"
	"rebuilder/jobs"
	"rebuilder/logger"
	"rebuilder/registry"
	"rebuilder/resources"
	"rebuilder/k8s"
	"rebuilder/secret"
)

func main() {
	logger.Info("rebuilder.main run started", "version", "Development-version", "go", "Golang-version")
	if err := environ.Load(); err != nil {
		logger.Error("rebuilder.main", "environ.Load", err)
	}

	// initialize k8s config
	k8s.InitConfig()

	// read all kubernetes resources
	list := resources.GetList()

	// traverse the list and check if any base image is newer than the derived image
	for _, entry := range list.Items {
		logger.Info("rebuilder.main", "namespace", entry.Metadata.Namespace, "name", entry.Metadata.Name)

		user, passwd := secret.GetCredentials(entry.Spec.Base.Authenticated, entry.Spec.Base.SecretName)
		baseTime := registry.GetLastUpdate(entry.Spec.Base.Host, entry.Spec.Base.Type, entry.Spec.Base.Image, entry.Spec.Base.Tag, user, passwd)

		user, passwd = secret.GetCredentials(entry.Spec.Registry.Authenticated, entry.Spec.Registry.SecretName)
		derivedTime := registry.GetLastUpdate(entry.Spec.Registry.Host, entry.Spec.Registry.Type, entry.Spec.Registry.Image, entry.Spec.Registry.Tag, user, passwd)

		// if yes spawn a job to rebuild the derived image, or sync the image to the private registry
		if baseTime.After(derivedTime) {
			logger.Info("rebuilder.main", "base image", fmt.Sprintf("%s:%s", entry.Spec.Base.Image, entry.Spec.Base.Tag), "derived image", fmt.Sprintf("%s:%s", entry.Spec.Registry.Image, entry.Spec.Registry.Tag), "up-to-date", "NO")
			err := jobs.RunBuildJob(entry.Spec.Git, entry.Spec.Registry, user, passwd)
			if err != nil {
				logger.Error("rebuilder.main", "job error", err)
			} else {
				// execute any after actions, restart deployment f.e.
				if err = actions.RunActions(entry.Metadata.Namespace, entry.Spec.Actions); err != nil {
					logger.Error("rebuilder.main", "actions.RunActions error", err)
				}
			}
		} else {
			if err := actions.RestartNeeded(entry.Metadata.Namespace, entry.Spec.Actions, derivedTime); err != nil {
				logger.Error("rebuilder.main", "actions.RestartNeeded error", err)
			}
			logger.Info("rebuilder.main", "derived image", fmt.Sprintf("%s:%s", entry.Spec.Registry.Image, entry.Spec.Registry.Tag), "up-to-date", "OK")
		}
	}
	logger.Info("rebuilder.main run finished")
}
