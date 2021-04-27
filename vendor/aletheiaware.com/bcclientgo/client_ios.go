// +build ios

/*
 * Copyright 2020 Aletheia Ware LLC
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

package bcclientgo

import "os"

func (c *bcClient) Root() (string, error) {
	if c.root == "" {
		if _, ok := os.LookupEnv("ROOT_DIRECTORY"); !ok {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				os.Setenv("ROOT_DIRECTORY") = homeDir
			}
		}
		if _, ok := os.LookupEnv("CACHE_DIRECTORY"); !ok {
			cacheDir, err := os.UserCacheDir()
			if err == nil {
				os.Setenv("CACHE_DIRECTORY") = cacheDir
			}
		}
		root, err := rootDir()
		if err != nil {
			return "", err
		}
		c.root = root
	}
	return c.root, nil
}
