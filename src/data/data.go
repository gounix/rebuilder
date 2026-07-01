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

package data

import (
	"fmt"
	"sync"
	"time"
	"errors"
	"rebuilder/logger"
)

type (
	StatusT struct {
		SrcOK    bool      // is the src registry accessible
		DstOK    bool      // is the dst registry accessible
		BuildOK  bool      // did the build job fail
		ActionOK bool      // was the restart succesful
		Message  string    // last error message
	}
	ProjectT struct {
		Namespace        string    // of the rebuild.yaml
		Name             string    // of the rebuild.yaml
		BaseImage        string    // registry/namespace/name:tag
		TargetImage      string    // registry/namespace/name:tag
		Updated          bool      // was this image updated during the last run
		Status           StatusT   // detailed statusses
		Timestamp        time.Time // timestamp of last update
	}
	dataT struct {
		mu          sync.Mutex
		projects    []ProjectT
		initialized bool
	}
)

var data = dataT{ initialized: false }

func Put(namespace string, name string, baseImage string, targetImage string, updated bool, status StatusT) {
	data.mu.Lock()
	defer data.mu.Unlock()

	found := false
	for nr, _ := range data.projects {
		if data.projects[nr].Namespace == namespace && data.projects[nr].Name == name {
			data.projects[nr].BaseImage   = baseImage
			data.projects[nr].TargetImage = targetImage
			data.projects[nr].Updated     = updated
			data.projects[nr].Status      = status
			data.projects[nr].Timestamp   = time.Now()

			data.initialized = true
			found = true

			logger.Info("data.Put updated", 
				"project", fmt.Sprintf("%s/%s", namespace, name), 
				"baseImage", baseImage, "targetImage", targetImage, 
				"updated", updated, "SrcOK", status.SrcOK, "DstOK", status.DstOK, 
				"buildOK", status.BuildOK, "actionOK", status.ActionOK)
		}
	}
	if ! found {
		var newEntry ProjectT

		newEntry.Namespace   = namespace
		newEntry.Name        = name
		newEntry.BaseImage   = baseImage
		newEntry.TargetImage = targetImage
		newEntry.Updated     = updated
		newEntry.Status      = status
		newEntry.Timestamp   = time.Now()

		data.projects = append(data.projects, newEntry)
		data.initialized = true
		logger.Info("data.Put initialized", 
			"project", fmt.Sprintf("%s/%s", namespace, name), 
			"baseImage", baseImage, "targetImage", targetImage, 
			"updated", updated, "SrcOK", status.SrcOK, "DstOK", status.DstOK, 
			"buildOK", status.BuildOK, "actionOK", status.ActionOK)
	}
}

func Get() ([]ProjectT, error) {
	data.mu.Lock()
	defer data.mu.Unlock()

	if !data.initialized {
		logger.Info("data.Get all not initialized")
		return []ProjectT{}, errors.New("not initialized")
	}

	// all is OK now
	logger.Info("data.Get", "first project", data.projects[0].BaseImage, "nr projects", len(data.projects))
	return data.projects, nil
}

func Alive(interval int64) bool {
	//now := time.Now()
	//diff := now.Sub(data.timestamp)

	// consider the producer dead after it missed 2 intervals
	//isOK := diff.Seconds() < float64(2 * interval)
	//logger.Info("data.Alive", "age(seconds)", math.Floor(diff.Seconds()), "OK", isOK)

	//return isOK
	return true
}

