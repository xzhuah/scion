// Copyright 2019 Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trc

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/scrypto"
)

// Parsing errors with context.
const (
	// InvalidKeyType indicates an inexistent key type.
	InvalidKeyType = "invalid key type"
	// InvalidAttribute indicates an inexistent attribute.
	InvalidAttribute = "invalid attribute"
	// InvalidAttributesSize indicates invalid number of attributes in the attributes list.
	InvalidAttributesSize = "invalid attributes size"
	// DuplicateAttributes indicates attribute duplication in the attributes list.
	DuplicateAttributes = "duplicate attributes"
)

// Invariant errors
const (
	// AuthoritativeButNotCore indicates a primary AS that is authoritative but not core.
	AuthoritativeButNotCore = "authoritative but not core"
	// UnexpectedKey indicates that a primary AS has an excess key. Voting ASes must
	// have an online and offline key. Non-Voting ASes must not have an offline
	// key. Issuer ASes must have an online key. Core-only ASes must not have
	// any key.
	UnexpectedKey = "unexpected key"
	// MissingKey indicates that the primary AS is missing a key.
	MissingKey = "missing key"
	// InvalidPrimaryAS indicates an invalid primary AS entry.
	InvalidPrimaryAS = "invalid primary as entry"
)

// Parsing errors
var (
	// ErrAttributesNotSet indicates the attributes in a primary AS are not set.
	ErrAttributesNotSet = errors.New("attributes not set")
	// ErrKeysNotSet indicates the keys in a primary AS are not set.
	ErrKeysNotSet = errors.New("keys not set")
)

// PrimaryASes holds all primary ASes and maps them to their attributes and keys.
type PrimaryASes map[addr.AS]PrimaryAS

// ValidateInvariant ensures that the TRC invariant holds for the primary ASes.
func (p *PrimaryASes) ValidateInvariant() error {
	for as, primary := range *p {
		if err := primary.ValidateInvariant(); err != nil {
			return common.NewBasicError(InvalidPrimaryAS, err, "as", as)
		}
	}
	return nil
}

// WithAttribute returns all primary ASes with the given attribute.
func (p *PrimaryASes) WithAttribute(attribute Attribute) PrimaryASes {
	m := make(PrimaryASes)
	for as, primary := range *p {
		if primary.Is(attribute) {
			m[as] = primary
		}
	}
	return m
}

// Count counts all primary ASes with the given attribute.
func (p *PrimaryASes) Count(attribute Attribute) int {
	var c int
	for _, primary := range *p {
		if primary.Is(attribute) {
			c++
		}
	}
	return c
}

// primaryASAlias is necessary to avoid an infinite recursion when unmarshalling.
type primaryASAlias PrimaryAS

// PrimaryAS holds the attributes and keys of a primary AS.
type PrimaryAS struct {
	Attributes Attributes                  `json:"attributes"`
	Keys       map[KeyType]scrypto.KeyMeta `json:"keys"`
}

// Is returns true if the primary AS holds this property.
func (p *PrimaryAS) Is(a Attribute) bool {
	for _, v := range p.Attributes {
		if v == a {
			return true
		}
	}
	return false
}

// ValidateInvariant ensures that the TRC invariant holds for the primary AS.
func (p *PrimaryAS) ValidateInvariant() error {
	if err := p.Attributes.Validate(); err != nil {
		return err
	}
	if err := p.checkKeyExistence(IssuingKey, p.Is(Issuing)); err != nil {
		return err
	}
	isVoting := p.Is(Voting)
	if err := p.checkKeyExistence(OnlineKey, isVoting); err != nil {
		return err
	}
	if err := p.checkKeyExistence(OfflineKey, isVoting); err != nil {
		return err
	}
	return nil
}

func (p *PrimaryAS) checkKeyExistence(keyType KeyType, shouldExist bool) error {
	_, ok := p.Keys[keyType]
	if ok && !shouldExist {
		return common.NewBasicError(UnexpectedKey, nil, "key_type", keyType)
	}
	if !ok && shouldExist {
		return common.NewBasicError(MissingKey, nil, "key_type", keyType)
	}
	return nil
}

// UnmarshalJSON checks that all fields are set.
func (p *PrimaryAS) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode((*primaryASAlias)(p)); err != nil {
		return err
	}
	return p.checkAllSet()
}

func (p *PrimaryAS) checkAllSet() error {
	switch {
	case p.Attributes == nil:
		return ErrAttributesNotSet
	case (p.Is(Voting) || p.Is(Issuing)) && p.Keys == nil:
		return ErrKeysNotSet
	}
	return nil
}

var _ json.Marshaler = (*Attributes)(nil)
var _ json.Unmarshaler = (*Attributes)(nil)

// Attributes holds all attributes of a primary AS.
type Attributes []Attribute

// Validate checks that the attributes list is valid.
func (t *Attributes) Validate() error {
	if len(*t) > 4 || len(*t) <= 0 {
		return common.NewBasicError(InvalidAttributesSize, nil, "len", len(*t))
	}
	var core, authoritative bool
	for i := 0; i < len(*t); i++ {
		core = core || (*t)[i] == Core
		authoritative = authoritative || (*t)[i] == Authoritative
		for j := i + 1; j < len(*t); j++ {
			if (*t)[i] == (*t)[j] {
				return common.NewBasicError(DuplicateAttributes, nil, "attribute", (*t)[i])
			}
		}
	}
	if authoritative && !core {
		return common.NewBasicError(AuthoritativeButNotCore, nil)
	}
	return nil
}

// MarshalJSON validates the attributes list during marshaling. It has to be a
// value receiver.
func (t Attributes) MarshalJSON() ([]byte, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(([]Attribute)(t))
}

// UnmarshalJSON validates the attributes list during parsing.
func (t *Attributes) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, (*[]Attribute)(t)); err != nil {
		return err
	}
	return t.Validate()
}

const (
	// Authoritative indicates an authoritative AS.
	Authoritative Attribute = "authoritative"
	// Core indicates a core AS.
	Core Attribute = "core"
	// Issuing indicates an issuing AS.
	Issuing Attribute = "issuing"
	// Voting indicates a voting AS. A voting AS must also be a core AS.
	Voting Attribute = "voting"
)

// Attribute indicates the capability of a primary AS.
type Attribute string

// UnmarshalText checks that the attribute is valid. It can either be
// "authoritative", "core", "issuing", or "voting".
func (t *Attribute) UnmarshalText(b []byte) error {
	switch Attribute(b) {
	case Authoritative:
		*t = Authoritative
	case Issuing:
		*t = Issuing
	case Voting:
		*t = Voting
	case Core:
		*t = Core
	default:
		return common.NewBasicError(InvalidAttribute, nil, "input", string(b))
	}
	return nil
}

const (
	IssuingKeyJSON = "issuing"
	OnlineKeyJSON  = "online"
	OfflineKeyJSON = "offline"
)

const (
	unknownKey KeyType = iota
	// IssuingKey is the issuing key type.
	IssuingKey
	// OnlineKey is the online key type.
	OnlineKey
	// OfflineKey is the offline key type.
	OfflineKey
)

// KeyType indicates the type of the key authenticated by the TRC.
//
// Because KeyType is used as a map key, it cannot be a string type. (see:
// https://github.com/golang/go/issues/33298)
type KeyType int

// UnmarshalText allows KeyType to be used as a map key and do validation when parsing.
func (t *KeyType) UnmarshalText(b []byte) error {
	switch string(b) {
	case OnlineKeyJSON:
		*t = OnlineKey
	case OfflineKeyJSON:
		*t = OfflineKey
	case IssuingKeyJSON:
		*t = IssuingKey
	default:
		return common.NewBasicError(InvalidKeyType, nil, "input", string(b))
	}
	return nil

}

// MarshalText is implemented to allow KeyType to be used as JSON map key. This
// must be a value receiver in order for KeyType fields in a struct to marshal
// correctly.
func (t KeyType) MarshalText() ([]byte, error) {
	switch t {
	case OnlineKey:
		return []byte(OnlineKeyJSON), nil
	case OfflineKey:
		return []byte(OfflineKeyJSON), nil
	case IssuingKey:
		return []byte(IssuingKeyJSON), nil
	}
	return nil, common.NewBasicError(InvalidKeyType, nil, "key_type", int(t))
}
