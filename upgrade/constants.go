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

// UpgradeName is the name of this specific software upgrade used on-chain.
const UpgradeName = "stratum"

// UpgradeASCII is the ASCII art shown to node operators upon successful upgrade.
const UpgradeASCII = `

	███████╗████████╗██████╗  █████╗ ████████╗██╗   ██╗███╗   ███╗
	██╔════╝╚══██╔══╝██╔══██╗██╔══██╗╚══██╔══╝██║   ██║████╗ ████║
	███████╗   ██║   ██████╔╝███████║   ██║   ██║   ██║██╔████╔██║
	╚════██║   ██║   ██╔══██╗██╔══██║   ██║   ██║   ██║██║╚██╔╝██║
	███████║   ██║   ██║  ██║██║  ██║   ██║   ╚██████╔╝██║ ╚═╝ ██║
	╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝    ╚═════╝ ╚═╝     ╚═╝

`

// TestnetChainID is the Chain ID of the Noble testnet.
const TestnetChainID = "grand-1"

// TestnetHyperlaneDomain is the Hyperlane domain of the Noble testnet.
// Generated from: console.log(parseInt('0x'+Buffer.from('GRAN').toString('hex')))
// We truncate "GRAND" to "GRAN" to not exceed the uint32 maximum.
const TestnetHyperlaneDomain = 1196573006

// MainnetChainID is the Chain ID of the Noble mainnet.
const MainnetChainID = "noble-1"

// MainnetHyperlaneDomain is the Hyperlane domain of the Noble mainnet.
// Generated from: console.log(parseInt('0x'+Buffer.from('NOBL').toString('hex')))
// We truncate "NOBLE" to "NOBL" to not exceed the uint32 maximum.
const MainnetHyperlaneDomain = 1313817164
