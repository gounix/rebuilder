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

package dockerregistry

import (
	"fmt"
	"time"
	"rebuilder/logger"
	"rebuilder/jsonreq"
)

const (
	manifestUrlPattern = "https://%s/v2/%s/manifests/%s"
	blobUrlPattern     = "https://%s/v2/%s/blobs/%s"
)

type (
	ConfigT struct {
		Digest string `json:"digest"`
	}
	ManifestT struct {
		Config ConfigT `json:"config"`
	}
	BlobT struct {
		Created time.Time `json:"created"`
	}
)

func getDigest(host string, repo string, tag string) (string, error) {
	var dat ManifestT

	url := fmt.Sprintf(manifestUrlPattern, host, repo, tag)
	logger.Info("dockerregistry.getDigest", "url", url)

	if err := jsonreq.GetJsonResp(url, "", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json", &dat); err != nil {
                logger.Error("dockerregistry.getDigest", "err", err)
                return "", err
        }

	logger.Info("dockerregistry.getDigest", "repo", repo, "tag", tag, "digest", dat.Config.Digest)
	return dat.Config.Digest, nil
}

func GetLastUpdate(host string, repo string, tag string) time.Time {
	var dat BlobT

	digest, err := getDigest(host , repo , tag )
	if err != nil {
		logger.Error("dockerregistry.GetLastUpdate", "getDigest error", err)
		return time.Time{}
	}
	url := fmt.Sprintf(blobUrlPattern, host, repo, digest)
	logger.Info("dockerregistry.GetLastUpdate", "url", url)

	if err := jsonreq.GetJsonResp(url, "", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json", &dat); err != nil {
                logger.Error("dockerregistry.GetLastUpdate", "err", err)
                return time.Time{}
        }

	logger.Info("dockerregistry.GetLastUpdate", "repo", repo, "tag", tag, "created", dat.Created)

	//checkVersions(host , repo , tag )
	return dat.Created
}

