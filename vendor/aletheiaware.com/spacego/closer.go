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

package spacego

import "io"

type Closer struct {
	writer io.Writer
	closer func() error
}

func NewCloser(writer io.Writer, closer func() error) io.WriteCloser {
	return &Closer{
		writer: writer,
		closer: closer,
	}
}

func (c *Closer) Write(p []byte) (n int, err error) {
	return c.writer.Write(p)
}

func (c *Closer) Close() error {
	if f := c.closer; f != nil {
		return f()
	}
	return nil
}
