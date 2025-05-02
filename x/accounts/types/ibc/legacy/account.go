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

var _ cryptotypes.PubKey = &ForwardingPubKey{}

func (fpk *ForwardingPubKey) String() string {
	return fmt.Sprintf("PubKeyForwarding{%X}", fpk.Key)
}

func (fpk *ForwardingPubKey) Address() cryptotypes.Address { return fpk.Key }

func (fpk *ForwardingPubKey) Bytes() []byte { return fpk.Key }

func (*ForwardingPubKey) VerifySignature(_ []byte, _ []byte) bool {
	panic("PubKeyForwarding.VerifySignature should never be invoked")
}

func (fpk *ForwardingPubKey) Equals(other cryptotypes.PubKey) bool {
	if _, ok := other.(*ForwardingPubKey); !ok {
		return false
	}

	return bytes.Equal(fpk.Bytes(), other.Bytes())
}

func (*ForwardingPubKey) Type() string { return "forwarding" }
