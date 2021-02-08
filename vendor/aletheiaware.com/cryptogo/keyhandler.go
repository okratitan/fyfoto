/*
 * Copyright 2019 Aletheia Ware LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cryptogo

import (
	"encoding/base64"
	"github.com/golang/protobuf/proto"
	"log"
	"net/http"
	"time"
)

type KeyShareStore map[string]*KeyShare

func KeyShareHandler(keys KeyShareStore, timeout time.Duration) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL.Path)
		switch r.Method {
		case "GET":
			name := ""
			if results, ok := r.URL.Query()["name"]; ok && len(results) > 0 {
				name = results[0]
			}
			log.Println("Name", name)
			if k, ok := keys[name]; ok {
				data, err := proto.Marshal(k)
				if err != nil {
					log.Println(err)
					return
				}
				count, err := w.Write(data)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println("Wrote KeyShare", count, "bytes")
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case "POST":
			r.ParseForm()
			log.Println("Request", r)
			name := r.Form["name"]
			log.Println("Name", name)
			publicKey := r.Form["publicKey"]
			log.Println("PublicKey", publicKey)
			publicKeyFormat := r.Form["publicKeyFormat"]
			log.Println("PublicKeyFormat", publicKeyFormat)
			privateKey := r.Form["privateKey"]
			log.Println("PrivateKey", privateKey)
			privateKeyFormat := r.Form["privateKeyFormat"]
			log.Println("PrivateKeyFormat", privateKeyFormat)
			password := r.Form["password"]
			log.Println("Password", password)

			if len(name) > 0 && len(publicKey) > 0 && len(publicKeyFormat) > 0 && len(privateKey) > 0 && len(privateKeyFormat) > 0 && len(password) > 0 {
				n := name[0]
				publicKey, err := base64.RawURLEncoding.DecodeString(publicKey[0])
				if err != nil {
					log.Println(err)
					return
				}
				pubFormatValue, ok := PublicKeyFormat_value[publicKeyFormat[0]]
				if !ok {
					log.Println("Unrecognized Public Key Format")
					return
				}
				pubFormat := PublicKeyFormat(pubFormatValue)
				privateKey, err := base64.RawURLEncoding.DecodeString(privateKey[0])
				if err != nil {
					log.Println(err)
					return
				}
				privFormatValue, ok := PrivateKeyFormat_value[privateKeyFormat[0]]
				if !ok {
					log.Println("Unrecognized Private Key Format")
					return
				}
				privFormat := PrivateKeyFormat(privFormatValue)
				password, err := base64.RawURLEncoding.DecodeString(password[0])
				if err != nil {
					log.Println(err)
					return
				}
				keys[n] = &KeyShare{
					Name:          n,
					PublicKey:     publicKey,
					PublicFormat:  pubFormat,
					PrivateKey:    privateKey,
					PrivateFormat: privFormat,
					Password:      password,
				}
				go func() {
					// Delete mapping after timeout
					time.Sleep(timeout)
					log.Println("Expiring Keys", n)
					delete(keys, n)
				}()
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		default:
			log.Println("Unsupported method", r.Method)
		}
	}
}
