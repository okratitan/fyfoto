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

type MetaFilter interface {
	Filter(*Meta) bool
}

func NewNameFilter(names ...string) MetaFilter {
	return &nameFilter{
		names: names,
	}
}

type nameFilter struct {
	names []string
}

func (f *nameFilter) Filter(meta *Meta) bool {
	for _, value := range f.names {
		if meta.Name == value {
			return true
		}
	}
	return false
}

func NewTypeFilter(types ...string) MetaFilter {
	return &typeFilter{
		types: types,
	}
}

type typeFilter struct {
	types []string
}

func (f *typeFilter) Filter(meta *Meta) bool {
	for _, value := range f.types {
		if meta.Type == value {
			return true
		}
	}
	return false
}

type TagFilter interface {
	Filter(*Tag) bool
}

func NewTagFilter(tags ...string) TagFilter {
	return &tagFilter{
		tags: tags,
	}
}

type tagFilter struct {
	tags []string
}

func (f *tagFilter) Filter(tag *Tag) bool {
	for _, value := range f.tags {
		if tag.Value == value {
			return true
		}
	}
	return false
}
