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

package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"rebuilder/jsonreq"
	"rebuilder/logger"
	"strings"
	"time"
)

const (
	checkAuthUrlPattern = "https://%s/v2/"
	getTokenUrlPattern  = "%s?service=%s&scope=repository:%s:pull"
	manifestUrlPattern  = "https://%s/v2/%s/manifests/%s"
	blobUrlPattern      = "https://%s/v2/%s/blobs/%s"
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
		MediaType string      `json:"mediaType"`
		Manifest  []ManifestT `json:"manifests"`
	}
	ConfigT struct {
		Digest string `json:"digest"`
	}
	SingleT struct {
		MediaType string  `json:"mediaType"`
		Config    ConfigT `json:"config"`
	}
	BlobT struct {
		Created time.Time `json:"created"`
	}
)

func getValueFromString(str string, substr string) string {

	startPos := strings.Index(str, substr)

	// including starting quote
	subStartPos := startPos + len(substr) + 1
	endPos := strings.Index(str[subStartPos:], "\"")
	endPos += subStartPos

	found := str[subStartPos:endPos]
	return found
}

// parse header and split in realm + service
// www-authenticate: Bearer realm="https://auth.docker.io/token",service="registry.docker.io"
// www-authenticate: Bearer realm="https://quay.io/v2/auth",service="quay.io"
// www-authenticate: Bearer realm="https://ghcr.io/token",service="ghcr.io",scope="repository:user/image:pull"
func getRealmService(header string) (string, string) {
	logger.Info("registry.getRealmService", "header", header)

	realm := getValueFromString(header, "realm=")
	service := getValueFromString(header, "service=")
	logger.Info("registry.getRealmService", "realm", realm, "service", service)

	return realm, service
}

func checkAuth(host string, repo string, user string, passwd string) (string, string) {

	// first check the v2 endpoint tot see if authentication is needed
	url := fmt.Sprintf(checkAuthUrlPattern, host)
	logger.Info("registry.checkAuth", "url", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if user != "" {
		logger.Info("registry.checkAuth", "user", user)
		req.SetBasicAuth(user, passwd)
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("registry.checkAuth", "http.Get error", err)
		return "", ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		logger.Info("registry.checkAuth no authentication needed", "status", resp.Status)
		return "", "" // no authentication needed
	}
	if resp.StatusCode != 401 {
		// something else
		logger.Error("registry.checkAuth", "status", resp.Status)
		return "", ""
	}
	// error 401, authentication needed
	// https://datatracker.ietf.org/doc/html/rfc6750#section-3
	authHeader := resp.Header.Get("www-authenticate")
	return getRealmService(authHeader)
}

// https://docs.docker.com/reference/api/registry/auth/
func getToken(realm string, service string, repo string, user string, passwd string) string {
	var dat TokenRespT

	url := fmt.Sprintf(getTokenUrlPattern, realm, service, repo)
	logger.Info("registry.getToken", "url", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if user != "" {
		logger.Info("registry.getToken", "user", user)
		req.SetBasicAuth(user, passwd)
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("registry.getToken", "http.Get error", err)
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logger.Info("registry.getToken", "status", resp.Status)
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("registry.getToken", "io.ReadAll error", err)
		return ""
	}

	err = json.Unmarshal(body, &dat)
	if err != nil {
		logger.Error("registry.getToken", "json.Unmarshal error", err)
		return ""
	}

	logger.Info("registry.getToken", "token(truncated)", dat.Token[:10], "expires_in", dat.ExpiresIn, "issued_at", dat.IssuedAt)
	return dat.Token
}

func getDigestFromImageIndex(host string, token string, repo string, tag string) (string, error) {
	var dat ManifestsT

	url := fmt.Sprintf(manifestUrlPattern, host, repo, tag)
	logger.Info("registry.getDigestFromImageIndex", "url", url)

	// application/vnd.docker.distribution.manifest.list.v2+json added for gcr.io
	// application/vnd.oci.image.manifest.v1+json for dockerhub
	if err := jsonreq.GetJsonResp(url, token, "application/vnd.docker.distribution.manifest.list.v2+json,application/vnd.oci.image.index.v1+json", &dat); err != nil {
		// not always present, so no error
		logger.Info("registry.getDigestFromImageIndex", "err", err)
		return "", err
	}
	// multi architecture manifest list
	if dat.MediaType != "application/vnd.oci.image.index.v1+json" &&
	   dat.MediaType != "application/vnd.docker.distribution.manifest.list.v2+json" {
		logger.Warn("registry.getDigestFromImageIndex", "MediaType", dat.MediaType)
	}

	for _, entry := range dat.Manifest {
		if entry.Platform.Architecture == "amd64" {
			logger.Info("registry.getDigestFromImageIndex returning", "digest", entry.Digest, 
			            "arch", entry.Platform.Architecture, "created", entry.Annotations.Created, 
				    "url", entry.Annotations.Url, "version", entry.Annotations.Version)
			return entry.Digest, nil
		}
	}
	logger.Error("registry.getDigestFromImageIndex return architecture not found")
	return "", errors.New("not found")
}

func getDigestFromManifest(host string, token string, repo string, digest string) (string, error) {
	var dat SingleT

	url := fmt.Sprintf(manifestUrlPattern, host, repo, digest)
	logger.Info("registry.getDigestFromManifest", "url", url)

	if err := jsonreq.GetJsonResp(url, token, "application/vnd.docker.distribution.manifest.v2+json,application/vnd.oci.image.manifest.v1+json", &dat); err != nil {
		logger.Error("registry.getDigestFromManifest", "err", err)
		return "", err
	}
	// docker manifest
	if dat.MediaType != "application/vnd.oci.image.manifest.v1+json" && 
	   dat.MediaType != "application/vnd.docker.distribution.manifest.v2+json" {
		logger.Warn("registry.getDigestFromManifest", "MediaType", dat.MediaType)
	}

	logger.Info("registry.getDigestFromManifest returning", "digest", dat.Config.Digest)
	return dat.Config.Digest, nil
}

func getBlob(host string, digest string, token string, repo string, tag string) (time.Time, error) {
	var dat BlobT

	url := fmt.Sprintf(blobUrlPattern, host, repo, digest)
	logger.Info("registry.getBlob", "url", url)

	if err := jsonreq.GetJsonResp(url, token, "application/vnd.oci.image.config.v1+json", &dat); err != nil {
		logger.Error("registry.getBlob", "err", err)
		return time.Time{}, err
	}

	logger.Info("registry.getBlob", "repo", repo, "tag", tag, "digest", digest, "created", dat.Created)

	return dat.Created, nil
}

func GetLastUpdate(host string, host_type string, repo string, tag string, user string, passwd string) (time.Time, error) {
	token := ""
	realm, service := checkAuth(host, repo, user, passwd)
	if realm != "" && service != "" {
		token = getToken(realm, service, repo, user, passwd)
	}
	digest1, err := getDigestFromImageIndex(host, token, repo, tag)
	if err != nil {
		// there is no image index manifest, try a normal manifest
		digest1 = tag
	}
	// get manifest for specific arch
	digest2, err := getDigestFromManifest(host, token, repo, digest1)
	if err != nil {
		logger.Error("registry.GetLastUpdate", "err", err)
		return time.Time{}, err
	}
	datum, err := getBlob(host, digest2, token, repo, tag)
	if err != nil {
		logger.Error("registry.GetLastUpdate", "err", err)
		return time.Time{}, err
	}

	return datum, nil
}
