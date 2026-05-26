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

package dockerhub

import (
	"fmt"
	"time"
	"errors"
	"rebuilder/logger"
	"rebuilder/jsonreq"
)

const (
	tokenPattern       = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull"
	manifestUrlPattern = "https://registry-1.docker.io/v2/%s/manifests/%s"
	blobUrlPattern     = "https://registry-1.docker.io/v2/%s/blobs/%s"
)

type (
	TokenRespT struct {
		Token       string    `json:"token"`
		AccessToken string    `json:"access_token"`
		ExpiresIn   int64     `json:"expires_in"`
		IssuedAt    time.Time `json:"issued_at"`
	}
	AnnotationsT struct {
		Created time.Time `json:"org.opencontainers.image.created"`
		Url     string    `json:"org.opencontainers.image.url"`
		Version string    `json:"org.opencontainers.image.version"`
	}
	PlatformT struct {
		Architecture string `json:"architecture"`
		Os           string `json:"os"`
	}
	ManifestT struct {
		Digest      string       `json:"digest"`
		Platform    PlatformT    `json:"platform"`
		Annotations AnnotationsT `json:"annotations"`
	}
	ManifestsT struct {
		Manifest []ManifestT `json:"manifests"`
	}
	ConfigT struct {
		Digest string `json:"digest"`
	}
	SingleT struct {
		Config ConfigT `json:"config"`
	}
	BlobT struct {
		Created time.Time `json:"created"`
	}
)

// https://docs.docker.com/reference/api/registry/auth/
func getToken(repo string) string {
	var dat TokenRespT

	url := fmt.Sprintf(tokenPattern, repo)
	logger.Info("dockerHub.getToken", "url", url)

	if err := jsonreq.GetJsonResp(url, "", "", &dat); err != nil {
                logger.Error("dockerHub.getToken", "err", err)
                return ""
        }

	logger.Info("dockerHub.getToken", "token", dat.Token[:10], "expires_in", dat.ExpiresIn, "issued_at", dat.IssuedAt)
	return dat.Token
}

func getDigestFromManifests(token string, repo string, tag string) (string, error) {
        var dat ManifestsT

        url := fmt.Sprintf(manifestUrlPattern, repo, tag)
	logger.Info("dockerHub.getDigestFromManifests", "url", url)

	if err := jsonreq.GetJsonResp(url, token, "application/vnd.oci.image.index.v1+json", &dat); err != nil {
                logger.Error("dockerHub.getDigestFromManifests", "err", err)
                return "", err
        }

	for _, entry := range dat.Manifest {
		if entry.Platform.Architecture == "amd64" {
			logger.Info("dockerHub.getDigestFromManifests returning", "digest", entry.Digest, "arch", entry.Platform.Architecture, "created", entry.Annotations.Created, "url", entry.Annotations.Url, "version", entry.Annotations.Version)
			return entry.Digest, nil
		}
	}
	logger.Error("dockerHub.getDigestFromManifests return not found")
        return "", errors.New("not found")
}

func getDigestFromSingle(token string, repo string, digest string) (string, error) {
        var dat SingleT

        url := fmt.Sprintf(manifestUrlPattern, repo, digest)
	logger.Info("dockerHub.getDigestFromSingle", "url", url)

	if err := jsonreq.GetJsonResp(url, token, "application/vnd.oci.image.manifest.v1+json", &dat); err != nil {
                logger.Error("dockerHub.getDigestFromSingle", "err", err)
                return "", err
        }

	logger.Info("dockerHub.getDigestFromSingle returning", "digest", dat.Config.Digest)
        return dat.Config.Digest, nil
}

func getBlob(digest string, token string, repo string, tag string) (time.Time, error) {
        var dat BlobT

        url := fmt.Sprintf(blobUrlPattern, repo, digest)
	logger.Info("dockerHub.getBlob", "url", url)

	if err := jsonreq.GetJsonResp(url, token, "application/vnd.oci.image.config.v1+json", &dat); err != nil {
                logger.Error("dockerHub.getBlob", "err", err)
                return time.Time{}, err
        }

	logger.Info("dockerHub.getBlob", "repo", repo, "tag", tag, "digest", digest, "created", dat.Created)

        return dat.Created, nil
}

func GetLastUpdate(host string, repo string, tag string) time.Time {

	token := getToken(repo)
	// get digest from master manifest
	digest1, _ := getDigestFromManifests(token, repo, tag)
	// get manifest for specific arch
	digest2, _ := getDigestFromSingle(token, repo, digest1)
	datum, _ := getBlob(digest2, token, repo, tag)

	//checkVersions(token, repo, tag)
	return datum
}
