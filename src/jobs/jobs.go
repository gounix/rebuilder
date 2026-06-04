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

package jobs

import (
	"context"
	"errors"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log/slog"
	"path/filepath"
	"rebuilder/environ"
	"rebuilder/logger"
	"rebuilder/resources"
	"time"
)

const sleepSeconds = 10

// newClient returns a kubernetes clientset from the given config path
// if path is empty it'll try to load the default kube config else error
func newClientSet() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	var kubeconfig string

	if environ.Env.Standalone {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			return nil, fmt.Errorf("could not find kubeconfig")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("error loading given file: %w", err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	return kubernetes.NewForConfig(config)
}

// createJobSpec returns a job object that can be applied to cluster
// It'll return the yaml example to k8s job object
func createJobSpec(name string, git resources.GitT, reg resources.RegistryT) *batchv1.Job {
	var (
		trueVal           = true
		zeroVal     int32 = 0
		usernameVal corev1.EnvVarSource
		passwordVal corev1.EnvVarSource
		env         []corev1.EnvVar
		authEnv     []corev1.EnvVar
	)

	// add current timestamp, as job name should be unique
	name = fmt.Sprintf("%s--%d", name, time.Now().UTC().UnixMilli())

	env = []corev1.EnvVar{
		// info from client-go applyconfigurations/internal/internal.go
		{Name: "GIT_HOST", Value: git.Host},
		{Name: "GIT_PROJECT", Value: git.Project},
		{Name: "GIT_USER", Value: git.User},
		{Name: "GIT_SUBDIR", Value: git.Dir},
		{Name: "GIT_TAG", Value: git.Tag},
		{Name: "GIT_SSH_KEY", Value: git.SshKeyName},
		{Name: "REGISTRY_HOST", Value: reg.Host},
		{Name: "REGISTRY_AUTHENTICATED", Value: fmt.Sprintf("%t", reg.Authenticated)},
	}
	objectRef := corev1.LocalObjectReference{Name: reg.SecretName}
	if reg.Authenticated == true {
		usernameSelector := corev1.SecretKeySelector{Key: "username", LocalObjectReference: objectRef}
		usernameVal = corev1.EnvVarSource{SecretKeyRef: &usernameSelector}

		passwordSelector := corev1.SecretKeySelector{Key: "password", LocalObjectReference: objectRef}
		passwordVal = corev1.EnvVarSource{SecretKeyRef: &passwordSelector}
		authEnv = []corev1.EnvVar{
			{Name: "REGISTRY_USER", ValueFrom: &usernameVal},
			{Name: "REGISTRY_PASSWORD", ValueFrom: &passwordVal},
		}
		env = append(env, authEnv...)
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            name,
							Image:           fmt.Sprintf("%s/%s:%s", environ.Env.BuilderRepo, environ.Env.BuilderImage, environ.Env.BuilderTag),
							ImagePullPolicy: "Always",
							Env:             env,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &trueVal,
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "varlibcontainers",
									MountPath: "/var/lib/containers",
								},
								corev1.VolumeMount{
									Name:      "ssh-key",
									MountPath: "/root/.ssh2",
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "varlibcontainers",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						corev1.Volume{
							Name: "ssh-key",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: git.SecretName,
									Items: []corev1.KeyToPath{{
										Key:  git.SshKeyName,
										Path: git.SshKeyName,
									},
									}},
							},
						},
					},
				},
			},
			BackoffLimit: &zeroVal,
		},
	}
}

func waitForJob(clientset *kubernetes.Clientset, jobName string) error {

	for true {
		job, err := clientset.BatchV1().Jobs(environ.Env.BuilderNamespace).Get(context.TODO(), jobName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if job.Status.Succeeded > 0 {
			logger.Info("jobs.waitForJob succeeded", "job", jobName)
			return nil // Job ran successfully
		}
		if job.Status.Failed > 0 {
			logger.Error("jobs.waitForJob failed", "job", jobName)
			return errors.New("Job failed")
		}
		if job.Status.Active == 0 {
			logger.Info("jobs.waitForJob not started", "job", jobName)
		} else {
			logger.Info("jobs.waitForJob still running", "job", jobName)
		}
		time.Sleep(sleepSeconds * time.Second)
	}
	return nil // unreachable
}

func RunBuildJob(git resources.GitT, reg resources.RegistryT) error {

	// get kubernetes clientset
	clientset, err := newClientSet()
	if err != nil {
		slog.Error("error getting k8s clientset: %w", err)
		return err
	}

	// get job spec
	job := createJobSpec("builder", git, reg)

	// create a client for default namespace
	jobClient := clientset.BatchV1().Jobs(environ.Env.BuilderNamespace)

	// trigger the job
	_, err = jobClient.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error creating job: %w", err)
	}

	slog.Info("Job has been created successfully", "name", job.Name)

	return waitForJob(clientset, job.ObjectMeta.Name)
}
