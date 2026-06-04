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

package jsonreq

import (
	//"fmt"
	"encoding/json"
	"io"
	"net/http"
	"rebuilder/logger"
	"strings"
)

func ratelimit(header http.Header) {

	// value in the form "100;w=21600"
	if limit := header.Get("Ratelimit-Limit"); limit != "" {
		endposLimit := strings.IndexAny(limit, ";")
		logger.Info("jsonreq.ratelimit", "ratelimit-limit", limit[:endposLimit])
	}
	if remaining := header.Get("Ratelimit-Remaining"); remaining != "" {
		endposRemaining := strings.IndexAny(remaining, ";")
		logger.Info("jsonreq.ratelimit", "ratelimit-remaining", remaining[:endposRemaining])
	}
}

func GetJsonResp(url string, token string, accept string, dat any) error {

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if accept != "" {
		req.Header.Add("accept", accept)
	}

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("jsonreq.getJsonResp", "client.do error", err)
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logger.Error("jsonreq.getJsonResp", "status", resp.Status)
		return err
	}

	// check if ratelimit header is present for dockerhub
	ratelimit(resp.Header)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("jsonreq.getJsonResp", "io.ReadAll error", err)
		return err
	}

	return json.Unmarshal(body, dat)
}
