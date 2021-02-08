// +build android

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

func (c *BCClient) GetRoot() (string, error) {
	if c.Root == "" {
		if _, ok := os.LookupEnv("ROOT_DIRECTORY"); !ok {
			os.Setenv("ROOT_DIRECTORY", os.Getenv("FILESDIR"))
		}
		if _, ok := os.LookupEnv("CACHE_DIRECTORY"); !ok {
			os.Setenv("CACHE_DIRECTORY", os.Getenv("TMPDIR"))
		}
		root, err := rootDir()
		if err != nil {
			return "", err
		}
		c.Root = root
	}
	return c.Root, nil
}

//func (c *BCClient) RecordCameraWriteToBC()
//func (c *BCClient) RecordMicrophoneWriteToBC()
//func (c *BCClient) RecordBatteryWriteToBC()
//func (c *BCClient) RecordSignalWriteToBC()
//func (c *BCClient) RecordLogWriteToBC()
