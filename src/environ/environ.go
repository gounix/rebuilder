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

package environ

import (
        "go-simpler.org/env"
	"rebuilder/logger"
	"os"
)

type EnvT struct {
	Standalone       bool   `env:"STANDALONE" default:false`
        BuilderImage     string `env:"BUILDER_IMAGE,required"`
        BuilderRepo      string `env:"BUILDER_REPO,required"`
        BuilderTag       string `env:"BUILDER_TAG,required"`
        BuilderNamespace string `env:"BUILDER_NAMESPACE,required"`
}

var Env EnvT

func Load() error {
	if err := env.Load(&Env, nil); err != nil {
                logger.Error("rebuilder/environ", "env.Load", err)
		os.Exit(1)
        }
	logger.Info("rebuilder.environ loaded environment", "STANDALONE", Env.Standalone, "BUILDER_REPO", Env.BuilderRepo, "BUILDER_IMAGE", Env.BuilderImage, "BUILDER_TAG", Env.BuilderTag, "BUILDER_NAMESPACE", Env.BuilderNamespace)
	return nil
}
