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

package frontend

import (
        "fmt"
        "net/http"
        "rebuilder/data"
        "rebuilder/environ"
        "rebuilder/logger"
)

const promHeader = `# HELP rebuilder_stats statistics of the rebuilder job
# TYPE rebuilder_stats gauge`

func logRequest(r *http.Request) {
        logger.Info("frontend.logRequest", "Host", r.Host, "Method", r.Method, "Url", r.URL.Path, "UserAgent", r.UserAgent())
}

func sendPromLines(w http.ResponseWriter, entry data.ProjectT) {
        var str string

        str = fmt.Sprintf("rebuilder_stats{name=\"%s/%s\",base=\"%s\", target=\"%s\",updated=\"%t\",buildSuccessful=\"%t\",actionSuccessful=\"%t\",timestamp=\"%s\"} 1\n", 
		entry.Namespace, entry.Name, entry.BaseImage, entry.TargetImage, entry.Updated, entry.BuildSuccessful, entry.ActionSuccessful, entry.Timestamp.Format("2006-01-02 15:04:05"))
        fmt.Fprintf(w, str)
        logger.Info("frontend.sendPromLines", "str", str)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {

        logRequest(r)
        data, err := data.Get()
	if err == nil {
		fmt.Fprintln(w, promHeader)
		for _, entry := range data {
			sendPromLines(w, entry)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
        logRequest(r)
        if data.Alive(1) {
                fmt.Fprintf(w, "OK")
        } else {
                w.WriteHeader(http.StatusNotFound)
        }
}

func Server() {
        http.HandleFunc("/metrics", metricsHandler)
        http.HandleFunc("/health", healthHandler)

        addr := fmt.Sprintf(":%d", environ.Env.Port)

	logger.Info("frontend.Server", "listen", addr)
        http.ListenAndServe(addr, nil)
}

