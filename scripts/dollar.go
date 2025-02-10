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

package main

import (
	"cosmossdk.io/math"

	dollartypes "dollar.noble.xyz/types"
)

func main() {
	index := math.LegacyOneDec()
	supply := math.ZeroInt()
	stats := dollartypes.Stats{
		TotalPrincipal:    math.ZeroInt(),
		TotalYieldAccrued: math.ZeroInt(),
	}

	// https://sepolia.etherscan.io/tx/0x7984b13be07b1a9efab3f703d3dcef95136ef8c343d105b88729ea033c36cfe3
	// https://www.mintscan.io/noble-testnet/tx/A9AD688F29B051B68D5A0DBDF52A42610EC0FB0F37ADCD91AC9784C15C611D1A
	// -> Mint 1 $USDN: Index = 1.036361460246, Total Principal = 964914, Total Yield Accrued = 0
	index = math.LegacyNewDec(1036361460246).QuoInt64(1e12)
	principal := math.LegacyNewDec(1_000_000).Quo(index).TruncateInt()
	supply = supply.Add(math.NewInt(1_000_000))

	stats.TotalPrincipal = stats.TotalPrincipal.Add(principal)

	// https://sepolia.etherscan.io/tx/0xa8dab049984caf08747ff01c79f80112b08dcce446f5f8a40a76227a3a8d1cb3
	// https://www.mintscan.io/noble-testnet/tx/507CA5B75F6DBC4BA296E23C4A646BFD90A1966F658A1C478915FC85F6B7478C
	// -> Update Index: Index = 1.036395670985, Total Principal = 964914, Total Yield Accrued = 32
	index = math.LegacyNewDec(1036395670985).QuoInt64(1e12)
	yield := index.MulInt(stats.TotalPrincipal).TruncateInt().Sub(supply)
	supply = supply.Add(yield)

	stats.TotalYieldAccrued = stats.TotalYieldAccrued.Add(yield)

	// https://sepolia.etherscan.io/tx/0x37073323d295d60b0275d1ab631d8c7b7f0af31f09f6fc43f23c076d5d53a1ed
	// https://www.mintscan.io/noble-testnet/tx/DC8B3D3CF494A9FE507931C0F092674A56BC82BDC1F0B7FE53658A1AD9148BA7
	// -> Update Index: Index = 1.037503387862, Total Principal = 964914, Total Yield Accrued = 1101
	index = math.LegacyNewDec(1037503387862).QuoInt64(1e12)
	yield = index.MulInt(stats.TotalPrincipal).TruncateInt().Sub(supply)
	supply = supply.Add(yield)

	stats.TotalYieldAccrued = stats.TotalYieldAccrued.Add(yield)

	// https://sepolia.etherscan.io/tx/0xcf3412eaa0beb24f8cb08ddcd9af43abf4034195b5108f70a6a7b95a79e596ae
	// Delivered via Jester in block #23130168
	// -> Update Index: Index = 1.037690277684, Total Principal = 964914, Total Yield Accrued = 1281
	index = math.LegacyNewDec(1037690277684).QuoInt64(1e12)
	yield = index.MulInt(stats.TotalPrincipal).TruncateInt().Sub(supply)
	supply = supply.Add(yield)

	stats.TotalYieldAccrued = stats.TotalYieldAccrued.Add(yield)

	// https://sepolia.etherscan.io/tx/0xa913bf0e9b5a2fa578cf28e001bb3ca277ba9dce673724ac4ccb7ce09fe8a0af
	// Delivered via Jester in block #23130243
	// -> Mint 1 $USDN: Index = 1.037690850231, Total Principal = 1928592, Total Yield Accrued = 1282
	index = math.LegacyNewDec(1037690850231).QuoInt64(1e12)
	yield = index.MulInt(stats.TotalPrincipal).TruncateInt().Sub(supply)
	supply = supply.Add(yield)
	principal = math.LegacyNewDec(1_000_000).Quo(index).TruncateInt()
	supply = supply.Add(math.NewInt(1_000_000))

	stats.TotalPrincipal = stats.TotalPrincipal.Add(principal)
	stats.TotalYieldAccrued = stats.TotalYieldAccrued.Add(yield)

	// https://www.mintscan.io/noble-testnet/tx/458B008D1614731AFC5E1795C2B7EA5CCF1025CE974BF36D7C4FEEA4747A6E33
	// https://sepolia.etherscan.io/tx/0xec6c95c7a34d57d62e8827181f174ead915a0423463a43296b41fdb6550f807b
	// -> Burn 1 $USDN: Index = 1.037690850231, Total Principal = 964914, Total Yield Accrued = 1282
	principal = math.LegacyNewDec(1_000_000).Quo(index).TruncateInt()
	supply = supply.Sub(math.NewInt(1_000_000))

	stats.TotalPrincipal = stats.TotalPrincipal.Sub(principal)

	// https://sepolia.etherscan.io/tx/0xf0756978c39ea02f13ff745d347b20fb11aa15b36a2200536d4701820465e035
	// Delivered via Jester in block #23192359
	// -> Update Index: Index = 1.037824814971, Total Principal = 964914, Total Yield Accrued = 1411
	index = math.LegacyNewDec(1037824814971).QuoInt64(1e12)
	yield = index.MulInt(stats.TotalPrincipal).TruncateInt().Sub(supply)
	supply = supply.Add(yield)

	stats.TotalYieldAccrued = stats.TotalYieldAccrued.Add(yield)
}
