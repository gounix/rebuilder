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

package ghcr


import (
	"fmt"
	"errors"
        "time"
	"rebuilder/logger"
	"rebuilder/jsonreq"
)

const (
	getTokenPattern    = "https://ghcr.io/token?scope=repository:%s:pull"
	manifestUrlPattern = "https://ghcr.io/v2/%s/manifests/%s"
	blobUrlPattern     = "https://ghcr.io/v2/%s/blobs/%s"
)

type (
	TokenRespT struct {
		Token       string    `json:"token"`
	}
	PlatformT struct {
		Architecture string `json:"architecture"`
		Os           string `json:"os"`
	}
	ManifestT struct {
		Digest      string       `json:"digest"`
		Platform    PlatformT    `json:"platform"`
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

func getToken(repo string) string {
	var dat TokenRespT

        url := fmt.Sprintf(getTokenPattern, repo)
        logger.Info("ghcr.getToken", "url", url)
	if err := jsonreq.GetJsonResp(url, "", "", &dat); err != nil {
		logger.Error("ghcr.getToken", "err", err)
                return ""
	}
        logger.Info("ghcr.getToken", "token", dat.Token)
        return dat.Token
}

func getDigestFromManifests(token string, repo string, tag string) (string, error) {
        var dat ManifestsT

        url := fmt.Sprintf(manifestUrlPattern, repo, tag)
        logger.Info("ghcr.getDigestFromManifests", "url", url)
	if err := jsonreq.GetJsonResp(url, token, "application/vnd.oci.image.index.v1+json", &dat); err != nil {
		logger.Error("ghcr.getDigestFromManifests", "err", err)
                return "", err
	}

        for _, entry := range dat.Manifest {
                if entry.Platform.Architecture == "amd64" {
                        logger.Info("ghcr.getDigestFromManifests returning", "digest", entry.Digest, "arch", entry.Platform.Architecture)
                        return entry.Digest, nil
                }
        }
        logger.Error("ghcr.getDigestFromManifests return not found")
        return "", errors.New("not found")
}

func getDigestFromSingle(token string, repo string, digest string) (string, error) {
        var dat SingleT

        url := fmt.Sprintf(manifestUrlPattern, repo, digest)
        logger.Info("ghcr.getDigestFromSingle", "url", url)

	if err := jsonreq.GetJsonResp(url, token, "application/vnd.docker.distribution.manifest.v2+json", &dat); err != nil {
		logger.Error("ghcr.getDigestFromSingle", "err", err)
                return "", err
	}

        logger.Info("ghcr.getDigestFromSingle returning", "digest", dat.Config.Digest)
        return dat.Config.Digest, nil
}

func getBlob(digest string, token string, repo string, tag string) (time.Time, error) {
        var dat BlobT

        url := fmt.Sprintf(blobUrlPattern, repo, digest)
        logger.Info("ghcr.getBlob", "url", url)

	if err := jsonreq.GetJsonResp(url, token, "application/vnd.oci.image.config.v1+json", &dat); err != nil {
		logger.Error("ghcr.getBlob", "err", err)
                return time.Time{}, err
	}

        logger.Info("ghcr.getBlob", "repo", repo, "tag", tag, "digest", digest, "created", dat.Created)
        return dat.Created, nil
}

func GetLastUpdate(host string, repo string, tag string) time.Time {
	token := getToken(repo)
	digest1, _ := getDigestFromManifests(token, repo, tag)
	digest2, _ := getDigestFromSingle(token, repo, digest1)
	datum, _ := getBlob(digest2, token, repo, tag)

	//checkVersions(token, repo, tag)
	return datum
}

