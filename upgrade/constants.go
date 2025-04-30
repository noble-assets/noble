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
const UpgradeName = "stratrum"

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

// HyperlaneDefaultISMs defines the default Hyperlane ISMs for external chains
// that Noble will support at launch via a routing ISM.
var HyperlaneDefaultISMs = map[string][]struct {
	Domain     uint32
	Name       string
	Validators []string
	Threshold  uint32
}{
	TestnetChainID: {
		{
			// https://docs.hyperlane.xyz/docs/reference/default-ism-validators#auroratestnet
			Domain: 1313161555,
			Name:   "auroratestnet",
			Validators: []string{
				"0xab1a2c76bf4cced43fde7bc1b5b57b9be3e7f937", // Abacus Works
			},
			Threshold: 1,
		},
		{
			// https://docs.hyperlane.xyz/docs/reference/default-ism-validators#hyperliquidevmtestnet
			Domain: 998,
			Name:   "hyperevmtestnet",
			Validators: []string{
				"0xea673a92a23ca319b9d85cc16b248645cd5158da", // Abacus Works
			},
			Threshold: 1,
		},
	},
	MainnetChainID: {
		{
			// https://docs.hyperlane.xyz/docs/reference/default-ism-validators#aurora
			Domain: 1313161554,
			Name:   "aurora",
			Validators: []string{
				"0x37105aec3ff37c7bb0abdb0b1d75112e1e69fa86", // Abacus Works
				"0xcf0211fafbb91fd9d06d7e306b30032dc3a1934f", // Merkly
				"0x4f977a59fdc2d9e39f6d780a84d5b4add1495a36", // Mitosis
			},
			Threshold: 2,
		},
		{
			// https://docs.hyperlane.xyz/docs/reference/default-ism-validators#hyperevm
			Domain: 999,
			Name:   "hyperevm",
			Validators: []string{
				"0x01be14a9eceeca36c9c1d46c056ca8c87f77c26f", // Abacus Works
				"0xcf0211fafbb91fd9d06d7e306b30032dc3a1934f", // Merkly
				"0x4f977a59fdc2d9e39f6d780a84d5b4add1495a36", // Mitosis
				"0x36f2bd8200ede5f969d63a0a28e654392c51a193", // Imperator
			},
			Threshold: 3,
		},
	},
}
