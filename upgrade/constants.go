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

package upgrade

import hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"

// UpgradeName is the name of this specific software upgrade used on-chain.
const UpgradeName = "ignition"

// UpgradeASCII is the ASCII art shown to node operators upon successful upgrade.
const UpgradeASCII = `

	██╗ ██████╗ ███╗   ██╗██╗████████╗██╗ ██████╗ ███╗   ██╗
	██║██╔════╝ ████╗  ██║██║╚══██╔══╝██║██╔═══██╗████╗  ██║
	██║██║  ███╗██╔██╗ ██║██║   ██║   ██║██║   ██║██╔██╗ ██║
	██║██║   ██║██║╚██╗██║██║   ██║   ██║██║   ██║██║╚██╗██║
	██║╚██████╔╝██║ ╚████║██║   ██║   ██║╚██████╔╝██║ ╚████║
	╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝

`

// DevnetChainID is the Chain ID of the Noble devnet.
const DevnetChainID = "duke-1"

// ApplayerDevnetChainID is the Chain ID of the Noble Applayer devnet.
const ApplayerDevnetChainID = 662532

// TestnetChainID is the Chain ID of the Noble testnet.
const TestnetChainID = "grand-1"

// ApplayerTestnetChainID is the Chain ID of the Noble Applayer testnet.
const ApplayerTestnetChainID = 662531

// MainnetChainID is the Chain ID of the Noble mainnet.
const MainnetChainID = "noble-1"

// ApplayerMainnetChainID is the Chain ID of the Noble Applayer mainnet.
const ApplayerMainnetChainID = 66253

// DefaultISM is the default Hyperlane Routing ISM ID on all Noble networks.
var DefaultISM, _ = hyperlaneutil.DecodeHexAddress("0x726f757465725f69736d00000000000000000000000000010000000000000000")
