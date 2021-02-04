// +build release

/*
 * Copyright 2021 Aletheia Ware LLC
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

package bcgo

import "os"

// If the release build tag was provided, and the LIVE environment variable was not explicitly set, then set it to true.
func init() {
	_, ok := os.LookupEnv(LIVE_FLAG)
	if !ok {
		os.Setenv(LIVE_FLAG, "true")
	}
}
