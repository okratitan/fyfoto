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

package storage

import (
	"aletheiaware.com/bcclientgo"
	"encoding/base64"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage/repository"
	"strings"
)

// ErrInvalidURI may be thrown when trying to parse a URI that is not supported by this repository.
var ErrInvalidURI = errors.New("invalid URI")

type BCRepository interface {
	repository.Repository
	repository.CustomURIRepository
	repository.CopyableRepository
	repository.HierarchicalRepository
	repository.ListableRepository
	repository.MovableRepository
	repository.WritableRepository
	Register()
}

type bcRepository struct {
	client bcclientgo.BCClient
}

func NewBCRepository(client bcclientgo.BCClient) BCRepository {
	return &bcRepository{
		client: client,
	}
}

func (r *bcRepository) CanList(u fyne.URI) (bool, error) {
	// TODO
	return false, fmt.Errorf("%s: Not Yet Implemented", "BCRepository.CanList")
}

func (r *bcRepository) CanRead(u fyne.URI) (bool, error) {
	// TODO
	return false, fmt.Errorf("%s: Not Yet Implemented", "BCRepository.CanRead")
}

func (r *bcRepository) CanWrite(u fyne.URI) (bool, error) {
	// TODO
	return false, fmt.Errorf("%s: Not Yet Implemented", "BCRepository.CanWrite")
}

func (r *bcRepository) Child(fyne.URI, string) (fyne.URI, error) {
	// TODO
	return nil, fmt.Errorf("%s: Not Yet Implemented", "BCRepository.Child")
}

func (r *bcRepository) Copy(fyne.URI, fyne.URI) error {
	// TODO
	return fmt.Errorf("%s: Not Yet Implemented", "BCRepository.Copy")
}

func (r *bcRepository) CreateListable(u fyne.URI) error {
	// TODO
	return fmt.Errorf("%s: Not Yet Implemented", "BCRepository.CreateListable")
}

func (r *bcRepository) Delete(u fyne.URI) error {
	// BC is indelible
	return repository.ErrOperationNotSupported
}

func (r *bcRepository) Destroy(string) {
	// Do nothing
}

func (r *bcRepository) Exists(u fyne.URI) (bool, error) {
	// TODO
	return false, fmt.Errorf("%s: Not Yet Implemented", "BCRepository.Exists")
}

func (r *bcRepository) List(u fyne.URI) ([]fyne.URI, error) {
	// TODO
	return nil, fmt.Errorf("%s: Not Yet Implemented", "BCRepository.List")
}

func (r *bcRepository) Move(fyne.URI, fyne.URI) error {
	// BC is immutable
	return repository.ErrOperationNotSupported
}

func (r *bcRepository) Parent(fyne.URI) (fyne.URI, error) {
	// TODO
	return nil, fmt.Errorf("%s: Not Yet Implemented", "BCRepository.Parent")
}

func (r *bcRepository) ParseURI(s string) (fyne.URI, error) {
	if strings.HasPrefix(s, ALIAS_SCHEME_PREFIX) {
		return NewAliasURI(strings.TrimPrefix(s, ALIAS_SCHEME_PREFIX)), nil
	}
	if !strings.HasPrefix(s, BC_SCHEME_PREFIX) {
		return nil, ErrInvalidURI
	}
	s = strings.TrimPrefix(s, BC_SCHEME_PREFIX)
	s = strings.TrimSuffix(s, "/")

	parts := strings.Split(s, "/")

	channel := parts[0]

	var blockhash []byte
	if len(parts) > 1 {
		bh, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}
		blockhash = bh
	}

	var recordhash []byte
	if len(parts) > 2 {
		rh, err := base64.RawURLEncoding.DecodeString(parts[2])
		if err != nil {
			return nil, err
		}
		recordhash = rh
	}

	if recordhash != nil && len(recordhash) > 0 {
		return NewRecordURI(channel, blockhash, recordhash), nil
	} else if blockhash != nil && len(blockhash) > 0 {
		return NewBlockURI(channel, blockhash), nil
	}
	return NewChannelURI(channel), nil
}

func (r *bcRepository) Reader(u fyne.URI) (fyne.URIReadCloser, error) {
	// TODO
	return nil, fmt.Errorf("%s: Not Yet Implemented", "BCRepository.Reader")
}

func (r *bcRepository) Register() {
	repository.Register(BC_SCHEME, r)
}

func (r *bcRepository) Writer(u fyne.URI) (fyne.URIWriteCloser, error) {
	// TODO
	return nil, fmt.Errorf("%s: Not Yet Implemented", "BCRepository.Writer")
}
