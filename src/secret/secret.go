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

package secret

import (
        "context"
	"strings"
        "encoding/json"
        "maps"
	"errors"
        //coreV1 "k8s.io/api/core/v1"
        metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"rebuilder/k8s"
	"rebuilder/logger"
	"rebuilder/environ"
)

func getSecretValue(namespace string, name string, key string) ([]byte, error) {

        logger.Info("secret.getSecretValue", "namespace", namespace, "name", name, "key", key)
	secretsClient := k8s.ClientSet.CoreV1().Secrets(namespace)

        secret, err := secretsClient.Get(context.TODO(), name, metaV1.GetOptions{})
        if err != nil {
                logger.Error("secret.getSecretValue", "secretsClient.Get", err)
		return []byte{}, err
        }

	logger.Info("secret.getSecretValue", "type", secret.Type)
        return secret.Data[key], nil
}

func getCredentialsFromSecret(namespace string, name string) (string, string, error) {
	var dat map[string]interface{}

	value, err := getSecretValue(namespace, name, ".dockerconfigjson")
        if err != nil {
                logger.Error("secret.GetCredentialsFromSecret", "getSecretValue", err)
		return "", "", err
        }

        err = json.Unmarshal(value, &dat)
        if err != nil {
                logger.Error("secret.GetCredentialsFromSecret", "json.Unmarshal", err)
		return "", "", err
        }

        // the auths key
        for auths := range maps.Keys(dat) {
                //fmt.Printf("auths key %s \n", k)
                newdat := dat[auths].(map[string]interface{})
                // the docker-server key
                for k2 := range maps.Keys(newdat) {
                        logger.Info("secret.GetCredentialsFromSecret found registry credentials", "registry", k2)
                        server := newdat[k2].(map[string]interface{})
			return server["username"].(string), server["password"].(string), nil
                }
        }
	return "", "", errors.New("not found")
}

func GetCredentials(authenticated bool, secretname string) (string, string) {
        var namespace, object string

        user, passwd := "", ""
        if authenticated {
                slice := strings.Split(secretname,"/")
                switch len(slice) {
                case 1:
                        namespace, object = environ.Env.RebuilderNamespace, slice[0]
                case 2:
                        namespace, object = slice[0], slice[1]
                default:
                        logger.Error("secret.getCredentials", "malformed secretname", secretname)
                        return "", ""
                }
                user, passwd, _ = getCredentialsFromSecret(namespace, object)
        }
        return user, passwd
}

