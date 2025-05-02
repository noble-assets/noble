// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 NASD Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"bytes"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

var _ cryptotypes.PubKey = &PubKey{}

func (pk *PubKey) String() string {
	return fmt.Sprintf("PubKeyAutoCCTP{%X}", pk.Key)
}

func (pk *PubKey) Address() cryptotypes.Address { return pk.Key }

func (pk *PubKey) Bytes() []byte { return pk.Key }

func (*PubKey) VerifySignature(_ []byte, _ []byte) bool {
	panic("PubKey.VerifySignature should never be invoked with AutoCCTP custom key")
}

func (pk *PubKey) Equals(other cryptotypes.PubKey) bool {
	if _, ok := other.(*PubKey); !ok {
		return false
	}

	return bytes.Equal(pk.Bytes(), other.Bytes())
}

func (*PubKey) Type() string { return "autocctp" }
