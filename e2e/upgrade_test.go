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

package e2e_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"strconv"
	"strings"
	"testing"

	"cosmossdk.io/math"
	dollartypes "dollar.noble.xyz/types"
	portaltypes "dollar.noble.xyz/types/portal"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/noble-assets/noble/e2e"
	wormholetypes "github.com/noble-assets/wormhole/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
	vaautils "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func TestChainUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	genesisVersion := "v7.0.0"

	upgrades := []e2e.ChainUpgrade{
		{
			Image:       e2e.GhcrImage("v8.0.5"),
			UpgradeName: "helium",
			PreUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet, icaTs *e2e.ICATestSuite) {
				icaAddr, err := e2e.RegisterICAAccount(ctx, icaTs)
				require.NoError(t, err, "failed to setup ICA account")
				require.NotEmpty(t, icaAddr, "ICA address should not be empty")

				// After successfully creating and querying the ICA account we need to update the test suite value for later usage.
				icaTs.IcaAddress = icaAddr

				// Assert initial balance of the ICA is correct.
				initBal, err := noble.BankQueryBalance(ctx, icaAddr, noble.Config().Denom)
				require.NoError(t, err, "failed to query bank balance")
				require.True(t, initBal.Equal(icaTs.InitBal), "invalid balance expected(%s), got(%s)", icaTs.InitBal, initBal)

				// Create and fund a user on Noble for use as the dst address in the bank transfer that we
				// compose below.
				users := interchaintest.GetAndFundTestUsers(t, ctx, "user", icaTs.InitBal, icaTs.Host)
				dstAddress := users[0].FormattedAddress()

				transferAmount := math.NewInt(1_000_000)

				fromAddress := sdk.MustAccAddressFromBech32(icaAddr)
				toAddress := sdk.MustAccAddressFromBech32(dstAddress)
				coin := sdk.NewCoin(icaTs.Host.Config().Denom, transferAmount)
				msgs := []sdk.Msg{banktypes.NewMsgSend(fromAddress, toAddress, sdk.NewCoins(coin))}

				icaTs.Msgs = msgs

				err = e2e.SendICATx(ctx, icaTs)
				require.NoError(t, err, "failed to send ICA tx")

				// Assert that the updated balance of the dst address is correct.
				expectedBal := icaTs.InitBal.Add(transferAmount)
				updatedBal, err := noble.BankQueryBalance(ctx, dstAddress, noble.Config().Denom)
				require.NoError(t, err, "failed to query bank balance")
				require.True(t, updatedBal.Equal(expectedBal), "invalid balance expected(%s), got(%s)", expectedBal, updatedBal)
			},
			PostUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet, icaTs *e2e.ICATestSuite) {
				msgSend, ok := icaTs.Msgs[0].(*banktypes.MsgSend)
				require.True(t, ok, "expected MsgSend, got %T", icaTs.Msgs[0])

				coin := msgSend.Amount[0]

				// Verify that the previously created ICA no longer works.
				err := e2e.SendICATx(ctx, icaTs)
				require.Error(t, err, "should have failed to send ICA tx after v8.0.0 upgrade")

				// Assert that the balance of the dst address has not updated.
				expectedBal := icaTs.InitBal.Add(coin.Amount)
				bal, err := noble.BankQueryBalance(ctx, msgSend.ToAddress, noble.Config().Denom)
				require.NoError(t, err, "failed to query bank balance")
				require.True(t, bal.Equal(expectedBal), "invalid balance expected(%s), got(%s)", expectedBal, bal)
			},
		},
		{
			Image:       e2e.LocalImages[0],
			UpgradeName: "argentum",
			PostUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet, icaTs *e2e.ICATestSuite) {
				msgSend, ok := icaTs.Msgs[0].(*banktypes.MsgSend)
				require.True(t, ok, "expected MsgSend, got %T", icaTs.Msgs[0])

				coin := msgSend.Amount[0]

				// The ICA tx we sent in the previous PostUpgrade handler will be relayed once the chain restarts with
				// the changes in v8.1 so we need to account for that when asserting the dst address bal is correct.
				expectedBal := icaTs.InitBal.Add(coin.Amount).Add(coin.Amount)
				bal, err := noble.BankQueryBalance(ctx, msgSend.ToAddress, noble.Config().Denom)
				require.NoError(t, err, "failed to query bank balance")
				require.True(t, bal.Equal(expectedBal), "invalid balance expected(%s), got(%s)", expectedBal, bal)

				// Verify that the previously created ICA works again with new txs as well.
				err = e2e.SendICATx(ctx, icaTs)
				require.NoError(t, err, "failed to send ICA tx")

				// Assert that the balance of the dst address is correct.
				expectedBal = bal.Add(coin.Amount)
				updatedBal, err := noble.BankQueryBalance(ctx, msgSend.ToAddress, noble.Config().Denom)
				require.NoError(t, err, "failed to query bank balance")
				require.True(t, updatedBal.Equal(expectedBal), "invalid balance expected(%s), got(%s)", expectedBal, updatedBal)

				require.NoError(t, ExecuteGuardianSetUpdates(t, ctx, noble.Validators[0], authority))

				require.NoError(t, ExecutePortalTransactions(t, ctx, noble.Validators[0], authority))
			},
		},
	}

	e2e.TestChainUpgrade(t, genesisVersion, upgrades, true)
}

// guardianSetUpdates contains the governance verified action approvals to be
// executed on mainnet to perform the Wormhole Guardain Set updates (0 -> 4).
//
// https://github.com/wormhole-foundation/wormhole/blob/3797ed082150e6d66c0dce3fea7f2848364af7d5/deployments/mainnet/guardianSetVAAs.csv
var guardianSetUpdates = map[int][]byte{
	1: common.FromHex("010000000001007ac31b282c2aeeeb37f3385ee0de5f8e421d30b9e5ae8ba3d4375c1c77a86e77159bb697d9c456d6f8c02d22a94b1279b65b0d6a9957e7d3857423845ac758e300610ac1d2000000030001000000000000000000000000000000000000000000000000000000000000000400000000000005390000000000000000000000000000000000000000000000000000000000436f7265020000000000011358cc3ae5c097b213ce3c81979e1b9f9570746aa5ff6cb952589bde862c25ef4392132fb9d4a42157114de8460193bdf3a2fcf81f86a09765f4762fd1107a0086b32d7a0977926a205131d8731d39cbeb8c82b2fd82faed2711d59af0f2499d16e726f6b211b39756c042441be6d8650b69b54ebe715e234354ce5b4d348fb74b958e8966e2ec3dbd4958a7cdeb5f7389fa26941519f0863349c223b73a6ddee774a3bf913953d695260d88bc1aa25a4eee363ef0000ac0076727b35fbea2dac28fee5ccb0fea768eaf45ced136b9d9e24903464ae889f5c8a723fc14f93124b7c738843cbb89e864c862c38cddcccf95d2cc37a4dc036a8d232b48f62cdd4731412f4890da798f6896a3331f64b48c12d1d57fd9cbe7081171aa1be1d36cafe3867910f99c09e347899c19c38192b6e7387ccd768277c17dab1b7a5027c0b3cf178e21ad2e77ae06711549cfbb1f9c7a9d8096e85e1487f35515d02a92753504a8d75471b9f49edb6fbebc898f403e4773e95feb15e80c9a99c8348d"),
	2: common.FromHex("01000000010d0012e6b39c6da90c5dfd3c228edbb78c7a4c97c488ff8a346d161a91db067e51d638c17216f368aa9bdf4836b8645a98018ca67d2fec87d769cabfdf2406bf790a0002ef42b288091a670ef3556596f4f47323717882881eaf38e03345078d07a156f312b785b64dae6e9a87e3d32872f59cb1931f728cecf511762981baf48303668f0103cef2616b84c4e511ff03329e0853f1bd7ee9ac5ba71d70a4d76108bddf94f69c2a8a84e4ee94065e8003c334e899184943634e12043d0dda78d93996da073d190104e76d166b9dac98f602107cc4b44ac82868faf00b63df7d24f177aa391e050902413b71046434e67c770b19aecdf7fce1d1435ea0be7262e3e4c18f50ddc8175c0105d9450e8216d741e0206a50f93b750a47e0a258b80eb8fed1314cc300b3d905092de25cd36d366097b7103ae2d184121329ba3aa2d7c6cc53273f11af14798110010687477c8deec89d36a23e7948feb074df95362fc8dcbd8ae910ac556a1dee1e755c56b9db5d710c940938ed79bc1895a3646523a58bc55f475a23435a373ecfdd0107fb06734864f79def4e192497362513171530daea81f07fbb9f698afe7e66c6d44db21323144f2657d4a5386a954bb94eef9f64148c33aef6e477eafa2c5c984c01088769e82216310d1827d9bd48645ec23e90de4ef8a8de99e2d351d1df318608566248d80cdc83bdcac382b3c30c670352be87f9069aab5037d0b747208eae9c650109e9796497ff9106d0d1c62e184d83716282870cef61a1ee13d6fc485b521adcce255c96f7d1bca8d8e7e7d454b65783a830bddc9d94092091a268d311ecd84c26010c468c9fb6d41026841ff9f8d7368fa309d4dbea3ea4bbd2feccf94a92cc8a20a226338a8e2126cd16f70eaf15b4fc9be2c3fa19def14e071956a605e9d1ac4162010e23fcb6bd445b7c25afb722250c1acbc061ed964ba9de1326609ae012acdfb96942b2a102a2de99ab96327859a34a2b49a767dbdb62e0a1fb26af60fe44fd496a00106bb0bac77ac68b347645f2fb1ad789ea9bd76fb9b2324f25ae06f97e65246f142df717f662e73948317182c62ce87d79c73def0dba12e5242dfc038382812cfe00126da03c5e56cb15aeeceadc1e17a45753ab4dc0ec7bf6a75ca03143ed4a294f6f61bc3f478a457833e43084ecd7c985bf2f55a55f168aac0e030fc49e845e497101626e9d9a5d9e343f00010000000000000000000000000000000000000000000000000000000000000004c1759167c43f501c2000000000000000000000000000000000000000000000000000000000436f7265020000000000021358cc3ae5c097b213ce3c81979e1b9f9570746aa5ff6cb952589bde862c25ef4392132fb9d4a42157114de8460193bdf3a2fcf81f86a09765f4762fd1107a0086b32d7a0977926a205131d8731d39cbeb8c82b2fd82faed2711d59af0f2499d16e726f6b211b39756c042441be6d8650b69b54ebe715e234354ce5b4d348fb74b958e8966e2ec3dbd4958a7cd66b9590e1c41e0b226937bf9217d1d67fd4e91f574a3bf913953d695260d88bc1aa25a4eee363ef0000ac0076727b35fbea2dac28fee5ccb0fea768eaf45ced136b9d9e24903464ae889f5c8a723fc14f93124b7c738843cbb89e864c862c38cddcccf95d2cc37a4dc036a8d232b48f62cdd4731412f4890da798f6896a3331f64b48c12d1d57fd9cbe7081171aa1be1d36cafe3867910f99c09e347899c19c38192b6e7387ccd768277c17dab1b7a5027c0b3cf178e21ad2e77ae06711549cfbb1f9c7a9d8096e85e1487f35515d02a92753504a8d75471b9f49edb6fbebc898f403e4773e95feb15e80c9a99c8348d"),
	3: common.FromHex("01000000020d00ce45474d9e1b1e7790a2d210871e195db53a70ffd6f237cfe70e2686a32859ac43c84a332267a8ef66f59719cf91cc8df0101fd7c36aa1878d5139241660edc0010375cc906156ae530786661c0cd9aef444747bc3d8d5aa84cac6a6d2933d4e1a031cffa30383d4af8131e929d9f203f460b07309a647d6cd32ab1cc7724089392c000452305156cfc90343128f97e499311b5cae174f488ff22fbc09591991a0a73d8e6af3afb8a5968441d3ab8437836407481739e9850ad5c95e6acfcc871e951bc30105a7956eefc23e7c945a1966d5ddbe9e4be376c2f54e45e3d5da88c2f8692510c7429b1ea860ae94d929bd97e84923a18187e777aa3db419813a80deb84cc8d22b00061b2a4f3d2666608e0aa96737689e3ba5793810ff3a52ff28ad57d8efb20967735dc5537a2e43ef10f583d144c12a1606542c207f5b79af08c38656d3ac40713301086b62c8e130af3411b3c0d91b5b50dcb01ed5f293963f901fc36e7b0e50114dce203373b32eb45971cef8288e5d928d0ed51cd86e2a3006b0af6a65c396c009080009e93ab4d2c8228901a5f4525934000b2c26d1dc679a05e47fdf0ff3231d98fbc207103159ff4116df2832eea69b38275283434e6cd4a4af04d25fa7a82990b707010aa643f4cf615dfff06ffd65830f7f6cf6512dabc3690d5d9e210fdc712842dc2708b8b2c22e224c99280cd25e5e8bfb40e3d1c55b8c41774e287c1e2c352aecfc010b89c1e85faa20a30601964ccc6a79c0ae53cfd26fb10863db37783428cd91390a163346558239db3cd9d420cfe423a0df84c84399790e2e308011b4b63e6b8015010ca31dcb564ac81a053a268d8090e72097f94f366711d0c5d13815af1ec7d47e662e2d1bde22678113d15963da100b668ba26c0c325970d07114b83c5698f46097010dc9fda39c0d592d9ed92cd22b5425cc6b37430e236f02d0d1f8a2ef45a00bde26223c0a6eb363c8b25fd3bf57234a1d9364976cefb8360e755a267cbbb674b39501108db01e444ab1003dd8b6c96f8eb77958b40ba7a85fefecf32ad00b7a47c0ae7524216262495977e09c0989dd50f280c21453d3756843608eacd17f4fdfe47600001261025228ef5af837cb060bcd986fcfa84ccef75b3fa100468cfd24e7fadf99163938f3b841a33496c2706d0208faab088bd155b2e20fd74c625bb1cc8c43677a0163c53c409e0c5dfa000100000000000000000000000000000000000000000000000000000000000000046c5a054d7833d1e42000000000000000000000000000000000000000000000000000000000436f7265020000000000031358cc3ae5c097b213ce3c81979e1b9f9570746aa5ff6cb952589bde862c25ef4392132fb9d4a42157114de8460193bdf3a2fcf81f86a09765f4762fd1107a0086b32d7a0977926a205131d8731d39cbeb8c82b2fd82faed2711d59af0f2499d16e726f6b211b39756c042441be6d8650b69b54ebe715e234354ce5b4d348fb74b958e8966e2ec3dbd4958a7cd15e7caf07c4e3dc8e7c469f92c8cd88fb8005a2074a3bf913953d695260d88bc1aa25a4eee363ef0000ac0076727b35fbea2dac28fee5ccb0fea768eaf45ced136b9d9e24903464ae889f5c8a723fc14f93124b7c738843cbb89e864c862c38cddcccf95d2cc37a4dc036a8d232b48f62cdd4731412f4890da798f6896a3331f64b48c12d1d57fd9cbe7081171aa1be1d36cafe3867910f99c09e347899c19c38192b6e7387ccd768277c17dab1b7a5027c0b3cf178e21ad2e77ae06711549cfbb1f9c7a9d8096e85e1487f35515d02a92753504a8d75471b9f49edb6fbebc898f403e4773e95feb15e80c9a99c8348d"),
	4: common.FromHex("01000000030d03d4a37a6ff4361d91714730831e9d49785f61624c8f348a9c6c1d82bc1d98cadc5e936338204445c6250bb4928f3f3e165ad47ca03a5d63111168a2de4576856301049a5df10464ea4e1961589fd30fc18d1970a7a2ffaad617e56a0f7777f25275253af7d10a0f0f2494dc6e99fc80e444ab9ebbbee252ded2d5dcb50cbf7a54bb5a01055f4603b553b9ba9e224f9c55c7bca3da00abb10abd19e0081aecd3b352be061a70f79f5f388ebe5190838ef3cd13a2f22459c9a94206883b739c90b40d5d74640006a8fade3997f650a36e46bceb1f609edff201ab32362266f166c5c7da713f6a19590c20b68ed3f0119cb24813c727560ede086b3d610c2d7a1efa66f655bad90900080f5e495a75ea52241c59d145c616bfac01e57182ad8d784cbcc9862ed3afb60c0983ccbc690553961ffcf115a0c917367daada8e60be2cbb8b8008bac6341a8c010935ab11e0eea28b87a1edc5ccce3f1fac25f75b5f640fe6b0673a7cd74513c9dc01c544216cf364cc9993b09fda612e0cd1ced9c00fb668b872a16a64ebb55d27010ab2bc39617a2396e7defa24cd7c22f42dc31f3c42ffcd9d1472b02df8468a4d0563911e8fb6a4b5b0ce0bd505daa53779b08ff660967b31f246126ed7f6f29a7e000bdb6d3fd7b33bdc9ac3992916eb4aacb97e7e21d19649e7fa28d2dd6e337937e4274516a96c13ac7a8895da9f91948ea3a09c25f44b982c62ce8842b58e20c8a9000d3d1b19c8bb000856b6610b9d28abde6c35cb7705c6ca5db711f7be96d60eed9d72cfa402a6bfe8bf0496dbc7af35796fc768da51a067b95941b3712dce8ae1e7010ec80085033157fd1a5628fc0c56267469a86f0e5a66d7dede1ad4ce74ecc3dff95b60307a39c3bfbeedc915075070da30d0395def9635130584f709b3885e1bdc0010fc480eb9ee715a2d151b23722b48b42581d7f4001fc1696c75425040bfc1ffc5394fe418adb2b64bd3dc692efda4cc408163677dbe233b16bcdabb853a20843301118ee9e115e1a0c981f19d0772b850e666591322da742a9a12cce9f52a5665bd474abdd59c580016bee8aae67fdf39b315be2528d12eec3a652910e03cc4c6fa3801129d0d1e2e429e969918ec163d16a7a5b2c6729aa44af5dccad07d25d19891556a79b574f42d9adbd9e2a9ae5a6b8750331d2fccb328dd94c3bf8791ee1bfe85aa00661e99781981faea00010000000000000000000000000000000000000000000000000000000000000004fd4c6c55ec8dfd342000000000000000000000000000000000000000000000000000000000436f726502000000000004135893b5a76c3f739645648885bdccc06cd70a3cd3ff6cb952589bde862c25ef4392132fb9d4a42157114de8460193bdf3a2fcf81f86a09765f4762fd1107a0086b32d7a0977926a205131d8731d39cbeb8c82b2fd82faed2711d59af0f2499d16e726f6b211b39756c042441be6d8650b69b54ebe715e234354ce5b4d348fb74b958e8966e2ec3dbd4958a7cd15e7caf07c4e3dc8e7c469f92c8cd88fb8005a2074a3bf913953d695260d88bc1aa25a4eee363ef0000ac0076727b35fbea2dac28fee5ccb0fea768eaf45ced136b9d9e24903464ae889f5c8a723fc14f93124b7c738843cbb89e864c862c38cddcccf95d2cc37a4dc036a8d232b48f62cdd4731412f4890da798f6896a3331f64b48c12d1d57fd9cbe7081171aa1be1d36cafe3867910f99c09e347899c19c38192b6e7387ccd768277c17dab1b7a5027c0b3cf178e21ad2e77ae06711549cfbb1f9c7a9d8096e85e1487f35515d02a92753504a8d75471b9f49edb6fbebc898f403e4773e95feb15e80c9a99c8348d"),
}

// guardianSetAddresses contains the list of expected addresses for each mainnet Wormhole Guardian Set.
var guardianSetAddresses = map[uint32][]string{
	// https://github.com/wormhole-foundation/wormhole/blob/3797ed082150e6d66c0dce3fea7f2848364af7d5/ethereum/env/.env.ethereum.mainnet#L4
	0: {"0x58CC3AE5C097b213cE3c81979e1B9f9570746AA5"},
	// https://github.com/wormhole-foundation/wormhole-networks/blob/master/mainnetv2/guardianset/v1.prototxt
	1: {"0x58CC3AE5C097b213cE3c81979e1B9f9570746AA5", "0xfF6CB952589BDE862c25Ef4392132fb9D4A42157", "0x114De8460193bdf3A2fCf81f86a09765F4762fD1", "0x107A0086b32d7A0977926A205131d8731D39cbEB", "0x8C82B2fd82FaeD2711d59AF0F2499D16e726f6b2", "0x11b39756C042441BE6D8650b69b54EbE715E2343", "0x54Ce5B4D348fb74B958e8966e2ec3dBd4958a7cd", "0xeB5F7389Fa26941519f0863349C223b73a6DDEE7", "0x74a3bf913953D695260D88BC1aA25A4eeE363ef0", "0x000aC0076727b35FBea2dAc28fEE5cCB0fEA768e", "0xAF45Ced136b9D9e24903464AE889F5C8a723FC14", "0xf93124b7c738843CBB89E864c862c38cddCccF95", "0xD2CC37A4dc036a8D232b48f62cDD4731412f4890", "0xDA798F6896A3331F64b48c12D1D57Fd9cbe70811", "0x71AA1BE1D36CaFE3867910F99C09e347899C19C3", "0x8192b6E7387CCd768277c17DAb1b7a5027c0b3Cf", "0x178e21ad2E77AE06711549CFBB1f9c7a9d8096e8", "0x5E1487F35515d02A92753504a8D75471b9f49EdB", "0x6FbEBc898F403E4773E95feB15E80C9A99c8348d"},
	// https://github.com/wormhole-foundation/wormhole-networks/blob/master/mainnetv2/guardianset/v2.prototxt
	2: {"0x58CC3AE5C097b213cE3c81979e1B9f9570746AA5", "0xfF6CB952589BDE862c25Ef4392132fb9D4A42157", "0x114De8460193bdf3A2fCf81f86a09765F4762fD1", "0x107A0086b32d7A0977926A205131d8731D39cbEB", "0x8C82B2fd82FaeD2711d59AF0F2499D16e726f6b2", "0x11b39756C042441BE6D8650b69b54EbE715E2343", "0x54Ce5B4D348fb74B958e8966e2ec3dBd4958a7cd", "0x66B9590e1c41e0B226937bf9217D1d67Fd4E91F5", "0x74a3bf913953D695260D88BC1aA25A4eeE363ef0", "0x000aC0076727b35FBea2dAc28fEE5cCB0fEA768e", "0xAF45Ced136b9D9e24903464AE889F5C8a723FC14", "0xf93124b7c738843CBB89E864c862c38cddCccF95", "0xD2CC37A4dc036a8D232b48f62cDD4731412f4890", "0xDA798F6896A3331F64b48c12D1D57Fd9cbe70811", "0x71AA1BE1D36CaFE3867910F99C09e347899C19C3", "0x8192b6E7387CCd768277c17DAb1b7a5027c0b3Cf", "0x178e21ad2E77AE06711549CFBB1f9c7a9d8096e8", "0x5E1487F35515d02A92753504a8D75471b9f49EdB", "0x6FbEBc898F403E4773E95feB15E80C9A99c8348d"},
	// https://github.com/wormhole-foundation/wormhole-networks/blob/master/mainnetv2/guardianset/v3.prototxt
	3: {"0x58CC3AE5C097b213cE3c81979e1B9f9570746AA5", "0xfF6CB952589BDE862c25Ef4392132fb9D4A42157", "0x114De8460193bdf3A2fCf81f86a09765F4762fD1", "0x107A0086b32d7A0977926A205131d8731D39cbEB", "0x8C82B2fd82FaeD2711d59AF0F2499D16e726f6b2", "0x11b39756C042441BE6D8650b69b54EbE715E2343", "0x54Ce5B4D348fb74B958e8966e2ec3dBd4958a7cd", "0x15e7cAF07C4e3DC8e7C469f92C8Cd88FB8005a20", "0x74a3bf913953D695260D88BC1aA25A4eeE363ef0", "0x000aC0076727b35FBea2dAc28fEE5cCB0fEA768e", "0xAF45Ced136b9D9e24903464AE889F5C8a723FC14", "0xf93124b7c738843CBB89E864c862c38cddCccF95", "0xD2CC37A4dc036a8D232b48f62cDD4731412f4890", "0xDA798F6896A3331F64b48c12D1D57Fd9cbe70811", "0x71AA1BE1D36CaFE3867910F99C09e347899C19C3", "0x8192b6E7387CCd768277c17DAb1b7a5027c0b3Cf", "0x178e21ad2E77AE06711549CFBB1f9c7a9d8096e8", "0x5E1487F35515d02A92753504a8D75471b9f49EdB", "0x6FbEBc898F403E4773E95feB15E80C9A99c8348d"},
	// https://github.com/wormhole-foundation/wormhole-networks/blob/master/mainnetv2/guardianset/v4.prototxt
	4: {"0x5893B5A76c3f739645648885bDCcC06cd70a3Cd3", "0xfF6CB952589BDE862c25Ef4392132fb9D4A42157", "0x114De8460193bdf3A2fCf81f86a09765F4762fD1", "0x107A0086b32d7A0977926A205131d8731D39cbEB", "0x8C82B2fd82FaeD2711d59AF0F2499D16e726f6b2", "0x11b39756C042441BE6D8650b69b54EbE715E2343", "0x54Ce5B4D348fb74B958e8966e2ec3dBd4958a7cd", "0x15e7cAF07C4e3DC8e7C469f92C8Cd88FB8005a20", "0x74a3bf913953D695260D88BC1aA25A4eeE363ef0", "0x000aC0076727b35FBea2dAc28fEE5cCB0fEA768e", "0xAF45Ced136b9D9e24903464AE889F5C8a723FC14", "0xf93124b7c738843CBB89E864c862c38cddCccF95", "0xD2CC37A4dc036a8D232b48f62cDD4731412f4890", "0xDA798F6896A3331F64b48c12D1D57Fd9cbe70811", "0x71AA1BE1D36CaFE3867910F99C09e347899C19C3", "0x8192b6E7387CCd768277c17DAb1b7a5027c0b3Cf", "0x178e21ad2E77AE06711549CFBB1f9c7a9d8096e8", "0x5E1487F35515d02A92753504a8D75471b9f49EdB", "0x6FbEBc898F403E4773E95feB15E80C9A99c8348d"},
}

// ExecuteGuardianSetUpdates ensures that post Argentum upgrade, that the
// Wormhole Guardian Sets can be correctly updated and registered.
func ExecuteGuardianSetUpdates(t *testing.T, ctx context.Context, validator *cosmos.ChainNode, wallet ibc.Wallet) error {
	for index := 1; index <= 4; index++ {
		t.Logf("Registering Wormhole Guardian Set %d", index)

		_, err := validator.ExecTx(
			ctx, wallet.KeyName(),
			"wormhole", "submit-vaa",
			base64.StdEncoding.EncodeToString(guardianSetUpdates[index]),
		)
		if err != nil {
			return err
		}

		_, _, err = validator.ExecQuery(ctx, "wormhole", "guardian-set", strconv.Itoa(index))
		if err != nil {
			return err
		}
	}

	raw, _, err := validator.ExecQuery(ctx, "wormhole", "config")
	if err != nil {
		return err
	}
	var configRes wormholetypes.QueryConfigResponse
	err = jsonpb.Unmarshal(bytes.NewReader(raw), &configRes)
	if err != nil {
		return err
	}
	config := configRes.Config

	raw, _, err = validator.ExecQuery(ctx, "wormhole", "guardian-sets")
	if err != nil {
		return err
	}
	var res wormholetypes.QueryGuardianSetsResponse
	err = jsonpb.Unmarshal(bytes.NewReader(raw), &res)
	if err != nil {
		return err
	}

	for index, guardianSet := range res.GuardianSets {
		if index == config.GuardianSetIndex {
			require.Zerof(t, guardianSet.ExpirationTime, "guardian set %d must not expire", index)
		} else {
			require.NotZerof(t, guardianSet.ExpirationTime, "guardian set %d must expire eventually", index)
		}

		expected := guardianSetAddresses[index]
		require.Len(t, guardianSet.Addresses, len(expected))

		for _, bz := range guardianSet.Addresses {
			address := common.BytesToAddress(bz).Hex()
			require.Contains(t, expected, address)
		}
	}

	return nil
}

// ExecutePortalTransactions ensures that post Argentum upgrade, that the
// initial Noble Dollar Portal messages can be delivered.
func ExecutePortalTransactions(t *testing.T, ctx context.Context, validator *cosmos.ChainNode, wallet ibc.Wallet) error {
	// 1. Index Propagation, 1.034586303552
	// https://etherscan.io/tx/0x9ff1d42bb1c425b683b83cea70782066bc1cde47c9da7a843be67f5f483e5215
	// https://wormholescan.io/#/tx/0x9ff1d42bb1c425b683b83cea70782066bc1cde47c9da7a843be67f5f483e5215?network=Mainnet&view=overview
	vaa := "AQAAAAQNAPotO0uTdQhgxz7rJfKntivZulNYWHlT0sAS78ksAc0EJ0CWOoFAPh8fGXacMhB55OX2dg+wxdTmc9cBJxZNyroBAX6/fISRub2csRxJaXw6QTL2vleRBQ3xn0269H5d6pF3Mf4h/7s+EIrSGvPp2QCfF1OTC97ALEB5/UPMQxFNJlwBAn6No27IaD2LJsKObDpV2kSq0PusnviMeG+CusXAT35uToRdEnsoY8iT+vfssMXQ36fLj1kAmZ3WSrQUVkYtYMcBBkxLqFX01H/Zsp3Pe0TNT7E7iTR1dh2hbEJqzAKu6YMIUoXyvV9omqX3l/d1gBz2lzBHxWUu8uP4L08zjGjLLi4AB0W8JAJvYH/VSds6asV9adUNcev+WveV/03y2a9JrykDVlOATBuQnkrFllp6foaW9U0wO9ElelCS/r6+j3kp3QoBCE+4r/g6eD14Gq/JUV16aWfV9xcz+qViY6JEhwpavPzIZkYRp96Dse+tXiOdSAsRruQ2Oqbk5YEKJq1C1nhbFwAACWoUAr5CjTwpB2ngLajUw2BiOaGdrneCDgcXOGuqNtNQSxxAJT1Q88czeVLfYnqhfTtUY5qBVKNT9qo+8dCANTwBCsDDWFwvhSeiH7YhBm3Zdyf+TtMCBIFwXXiBmS8b9I0NBPD7+KLl+IxqL+LHKzyEdFSU8kP1kC7jLO2wYbbtk6MADddW7L5gfC5GfM1Bli8vTIB4kKgjBupAqGwPqof582pNeVc+cRTv8QczZYNgmlHicKU21YEoUFxFackMbgT1EB0BDrS43I3yaoYI6eHz5iiHmbNNAr8E4jmynE2NbYQFuZfOQt59vdDtuV1l4Z5pvn3BPBn7HuA/nNSpo4Pll+RWkocBD1thCw+USxN9zLqvYEGHkNGWiuCGDyBn93LXRAozikB5CAq2eLrOMnt4g/H4HX3Mpo1pseOHCDc3vojFHesZTtEBEIOn+5uezSoF8BpnAAbeFM81YFSQxuwgSBijI1PPqNDSdqJgQWKOsKuA2E0lA+OvMhOYQojxQhY6pgtlPiqBVQUBETSUDicCPJ7FHhcvfDceN7c2gk36uMJsTpKD0vvODTJ5DBhfiLiFEkGJaEGNoF/QJGZmqQTOGIxQRnt+OmyVg6gBZ8X5XwAAAAAAAgAAAAAAAAAAAAAAAMfdNyw544vxFFGrSoQntK44zvZEAAAAAAAAAAIBmUX/EAAAAAAAAAAAAAAAAIOugr1AVOgV+3sYnDnZzmcDaeoWAAAAAAAAAAAAAAAALoWVBroinBg/iYXVT+chCSP7m8oAUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA/PdC2Bphg9KcRn1GYrvZmF+qWxMADk0wSVQAAADw4iYsQA+pAAA="

	hash, err := validator.ExecTx(ctx, wallet.KeyName(), "dollar", "portal", "deliver", vaa)
	if err != nil {
		return err
	}

	index, err := QueryIndex(ctx, validator)
	if err != nil {
		return err
	}
	require.Equal(t, math.LegacyMustNewDecFromStr("1.034586303552"), index)

	tx, err := QueryTransaction(ctx, validator, hash)
	if err != nil {
		return err
	}

	for _, rawEvent := range tx.Events {
		switch rawEvent.Type {
		case "noble.dollar.v1.IndexUpdated":
			event, err := sdk.ParseTypedEvent(rawEvent)
			if err != nil {
				return err
			}

			indexUpdated := event.(*dollartypes.IndexUpdated)

			require.Equal(t, int64(1e12), indexUpdated.OldIndex)
			require.Equal(t, int64(1034586303552), indexUpdated.NewIndex)
			require.Equal(t, math.ZeroInt(), indexUpdated.TotalPrincipal)
			require.Equal(t, math.ZeroInt(), indexUpdated.YieldAccrued)
		}
	}

	// 2. Transfer of 1 $M for $USDN, Index = 1.034586859032
	// https://etherscan.io/tx/0x5a2063704dfe1e8379fac019dad77bcbef806f352e2bd853301a1d64ba194129
	// https://wormholescan.io/#/tx/0x5a2063704dfe1e8379fac019dad77bcbef806f352e2bd853301a1d64ba194129?network=Mainnet&view=overview
	vaa = "AQAAAAQNANmG1ElQ5Rlx2LJTiRyFo3g6Zd4uujUOU04PqKjAKGzUcH9IXZcp7YIOpS/+UrBer0lwTW/J4x42Za/usmUqQhgAAd0z6mePw50gNpUr2HPCPC3dPjjkBMnR3R6vv/tI4ztyIIATcpBl9GtZ1QRyJpV34Hv351NhZUNK+KlOubTpJhQBAs1r9ssSOgnR8Sh6VqFOWNsTCLmpql2fV+YOBZSAsE0MffB2pv1SnWR7unI51wQYLaI7GZUFOSwc9tpM9HvncM4AAzwafd/jCfCWhrTn02uNzoa9jR1O4CSTCZ7M/ZOLEwYfa2U/Da0PeudHbwZe5mLoVFETh8Gh0KFaUAFVV1xRQaoBBHqcJj46wqZEOtGTh3nnnRWw4llvyFu9jOtuDluVhQBZYnbJZj03PCBmCxKxkZvYgJa/ApYAi6w/pi+eq3nufBkABmdaDJvwX7PDwRisVQD1KU99y0ez971Fn+VJKqKveXzOKY3DCC/ytipKt8kIe8jq3eJ+ZpDnyQo2ZwKohVNFJzYBB8KGEoy9yb60EPWjXFB3Bj/qNzyj6j+Mn229AuVY/h6+Ii3Ivje1ZiNOyJYHdC/QN/9nRp5TkJ90Z0Dr5m19vKoBCWSmrEDu3+N4QHNT1yFmFFStOZtJZEt7CO5KEq+W4tU9dy5nYmmf9Le696F1ALoSLN/4yC/OsP1ycL9iX2SICmEBCpkv4eCKTRKvt7K1V/DNIzwF4HmnK9orrOPtfgF/S1RkId7DZFQ9NTZD3DgTic4iRIz3Lv4huHq0s9n8PVkhxlkADA4EGp8trEKdrkfg6SGRUHLZ7p/cVa59Q7QPFdcI1rdmQQ7+06RUxABeu+GfpzUXhUL3Jc5LJuULXW2/HSRQuDUADbWbkdC0ocJlyTdOiBaA1T6m0US0Wom1fM7v79lQs/b3FIX6+8qPJRCALbn9w9q7/qVD2mPu6s3YfZyNkwB1e9wAEPTBN58JqZc/xLL8V/y+I8uwQHGFRmsywyVoxXjGod53GwzPr3sMvT4+p/FUtjglA7q4BY4vLr3EbtwDtqOdbrEBETQEyjaCj1xDt4JZTZvckLqSR4CDWGKAYI8pp5RH871jPgbugxNugyKzAbxui436ifQdRTdD3EdZtkxBtD0NvosAZ8X69wAAAAAAAgAAAAAAAAAAAAAAAMfdNyw544vxFFGrSoQntK44zvZEAAAAAAAAAAMBmUX/EAAAAAAAAAAAAAAAAIOugr1AVOgV+3sYnDnZzmcDaeoWAAAAAAAAAAAAAAAALoWVBroinBg/iYXVT+chCSP7m8oAuwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAA/PdC2Bphg9KcRn1GYrvZmF+qWxMAeZlOVFQGAAAAAAAPQj8AAAAAAAAAAAAAAACGaiv05XLLzzfVBxp6WFA7+za+GwAAAAAAAAAAAAAAADj6JvVmIfFz80NjymQuRaJLkXjsD6kAKAAAAPDiLqYYAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAdXVzZG4AAA=="

	hash, err = validator.ExecTx(ctx, wallet.KeyName(), "dollar", "portal", "deliver", vaa)
	if err != nil {
		return err
	}

	totalSupply, err := QueryTotalSupply(ctx, validator)
	if err != nil {
		return err
	}
	require.Equal(t, math.NewInt(999999), totalSupply)

	tx, err = QueryTransaction(ctx, validator, hash)
	if err != nil {
		return err
	}

	for _, rawEvent := range tx.Events {
		switch rawEvent.Type {
		case "noble.dollar.portal.v1.MTokenReceived":
			event, err := sdk.ParseTypedEvent(rawEvent)
			if err != nil {
				return err
			}

			mTokenReceived := event.(*portaltypes.MTokenReceived)
			destinationToken := strings.TrimLeft(string(mTokenReceived.DestinationToken), "\u0000")
			sender := common.BytesToAddress(mTokenReceived.Sender[12:]).String()
			messageId := strings.ToUpper(common.Bytes2Hex(mTokenReceived.MessageId))

			require.Equal(t, uint32(vaautils.ChainIDEthereum), mTokenReceived.SourceChainId)
			require.Equal(t, "uusdn", destinationToken)
			require.Equal(t, "0xfcF742d81A6183D29c467d4662BBd9985faa5B13", sender)
			require.Equal(t, "noble18razdatxy8ch8u6rv09xgtj95f9ez78vvcwjau", mTokenReceived.Recipient)
			require.Equal(t, math.NewInt(999999), mTokenReceived.Amount)
			require.Equal(t, int64(1034586859032), mTokenReceived.Index)
			require.Equal(t, "3BD0CC34D4AEB8E09D125590D13AB7229F8F1BDB4B17F2A1E1AAAC017F9E4830", messageId)
		case "noble.dollar.v1.IndexUpdated":
			event, err := sdk.ParseTypedEvent(rawEvent)
			if err != nil {
				return err
			}

			indexUpdated := event.(*dollartypes.IndexUpdated)

			require.Equal(t, int64(1034586303552), indexUpdated.OldIndex)
			require.Equal(t, int64(1034586859032), indexUpdated.NewIndex)
			require.Equal(t, math.ZeroInt(), indexUpdated.TotalPrincipal)
			require.Equal(t, math.ZeroInt(), indexUpdated.YieldAccrued)
		}
	}

	// 3. Transfer of 1 $wM for $USDN, Index = 1.034586989732
	// https://etherscan.io/tx/0x517ca887c67f6c6741eda76e0855f30c025f7b7d68916a32e438debbc622c2d3
	// https://wormholescan.io/#/tx/0x517ca887c67f6c6741eda76e0855f30c025f7b7d68916a32e438debbc622c2d3?network=Mainnet&view=overview
	vaa = "AQAAAAQNABSDpxF5RvSH/ZAR1Kl3EGj9urzecO77UUdtJp5q8wNTFEonaKKr8auAqXTHgT2Y+wmsfiroJZs7U0IdOHCR3F4AAW593ttl+E9twoNH8B3009QWL/qBpTqwGohWwma1k5guDWxk54C9BV+z9qXgUSiO4R9JXk/jMZ/LseEHVq7fa/QBAu6Vate5VtptXU9fkRawTyYkHXv2chuTQtDKGPvl9pUeEwM4Xh38wMn+YfKkDJvvMcZ646nJbgrlaJ2+zIwRXqQABOe+RLyQQovE3YHTvVKOnXP7vMviIf8KCfHCuEUfFwsWLu+jBTHj+1vhS8b0mFVT6Ry6z7SlaqDgk11j6O8vSg0ABr5iCYhct8CVPNnRr2mpGO06bdNTFnnV2QZfDJrvDWRketW9vaV6YrGXYsDstLy6Go8Lf6x14at1iQ7FyyWA26sABx/Ts62kO+5tS8Yb6nJMRXX2uJmWadQ65GuTTqQlXeU8NTrax9RNxneE2L0hAEyD/PAAr46qL1LH7aPjsQOplCwBCfrCdWjpOKxMQdKV0cogVpFeEhjGmLchgEkrHIJdNCJHRyWVaVs4Eio9wuIH+Bzar/OxWNcLjl3oHpm+4IUpzyYBCkgkltgR5d53CwvIf8TtItTA7ZieYepzyOOZ1zskJNEzIX3jIPHbrQnrwC9GM1VHNXQCfViS6OK3VHoDCtaPz2UADJcykKZSVnMlTmsfZRmQcChIwE03CAzl07bez3Aep1JBViALRnlxtpMFZNVkKe2BW+gsYCHkIpQ4BVky0yk2Y2EBDRKCjUXDp/oeYp9Jt6pmH7mKhLLJu4jwupo8Cr6CkTEfIVbF/+I7Bc2/lvFr5OYvFYCiYK3lTBJ/ot2Yc8t7PAEBDwk4Lt93AQxuSkXt2EN9cF4IA0mxgpTRyOPFsGjJS9x0WOTk0IlIYGK6VV258Iwxlu2fqX/UMV3HeNpWzQSSz1EBECYEJiqHea+/GJIS+OR4zVzcRawukFvmwGoC/qFI1w30WGeCzFPAydzczmfy5t4UQMap268+ibQx/IJestp9D50AEbTH5phuh5N3BCgUGr6xHu5iTjRo1G7A+HPtAfYCYgNuJcrL5yphYXGdclBBtIDzenfVdsaDVPfiUuEaNV6JSGYBZ8X7VwAAAAAAAgAAAAAAAAAAAAAAAMfdNyw544vxFFGrSoQntK44zvZEAAAAAAAAAAQBmUX/EAAAAAAAAAAAAAAAAIOugr1AVOgV+3sYnDnZzmcDaeoWAAAAAAAAAAAAAAAALoWVBroinBg/iYXVT+chCSP7m8oAuwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAAAAAAAAAAAAAAA/PdC2Bphg9KcRn1GYrvZmF+qWxMAeZlOVFQGAAAAAAAPQkAAAAAAAAAAAAAAAACGaiv05XLLzzfVBxp6WFA7+za+GwAAAAAAAAAAAAAAADj6JvVmIfFz80NjymQuRaJLkXjsD6kAKAAAAPDiMKSkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAdXVzZG4AAA=="

	hash, err = validator.ExecTx(ctx, wallet.KeyName(), "dollar", "portal", "deliver", vaa)
	if err != nil {
		return err
	}

	totalSupply, err = QueryTotalSupply(ctx, validator)
	if err != nil {
		return err
	}
	require.Equal(t, math.NewInt(1999999), totalSupply)

	tx, err = QueryTransaction(ctx, validator, hash)
	if err != nil {
		return err
	}

	for _, rawEvent := range tx.Events {
		switch rawEvent.Type {
		case "noble.dollar.portal.v1.MTokenReceived":
			event, err := sdk.ParseTypedEvent(rawEvent)
			if err != nil {
				return err
			}

			mTokenReceived := event.(*portaltypes.MTokenReceived)
			destinationToken := strings.TrimLeft(string(mTokenReceived.DestinationToken), "\u0000")
			sender := common.BytesToAddress(mTokenReceived.Sender[12:]).String()
			messageId := strings.ToUpper(common.Bytes2Hex(mTokenReceived.MessageId))

			require.Equal(t, uint32(vaautils.ChainIDEthereum), mTokenReceived.SourceChainId)
			require.Equal(t, "uusdn", destinationToken)
			require.Equal(t, "0xfcF742d81A6183D29c467d4662BBd9985faa5B13", sender)
			require.Equal(t, "noble18razdatxy8ch8u6rv09xgtj95f9ez78vvcwjau", mTokenReceived.Recipient)
			require.Equal(t, math.NewInt(1000000), mTokenReceived.Amount)
			require.Equal(t, int64(1034586989732), mTokenReceived.Index)
			require.Equal(t, "6211441CEB201621C8B9273FA15CF9FA5059773BA00912CADF1990E066DFB34B", messageId)
		case "noble.dollar.v1.IndexUpdated":
			event, err := sdk.ParseTypedEvent(rawEvent)
			if err != nil {
				return err
			}

			indexUpdated := event.(*dollartypes.IndexUpdated)

			require.Equal(t, int64(1034586859032), indexUpdated.OldIndex)
			require.Equal(t, int64(1034586989732), indexUpdated.NewIndex)
			require.Equal(t, math.NewInt(966569), indexUpdated.TotalPrincipal)
			require.Equal(t, math.ZeroInt(), indexUpdated.YieldAccrued)
		}
	}

	return nil
}

// QueryIndex is a utility for querying the latest Noble Dollar index.
func QueryIndex(ctx context.Context, validator *cosmos.ChainNode) (math.LegacyDec, error) {
	raw, _, err := validator.ExecQuery(ctx, "dollar", "index")
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	var res dollartypes.QueryIndexResponse
	err = jsonpb.Unmarshal(bytes.NewReader(raw), &res)
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	return res.Index, nil
}

// QueryTransaction is a utility for querying a transaction response.
func QueryTransaction(ctx context.Context, validator *cosmos.ChainNode, hash string) (*sdk.TxResponse, error) {
	raw, _, err := validator.ExecQuery(ctx, "tx", hash)
	if err != nil {
		return nil, err
	}

	var res sdk.TxResponse
	err = jsonpb.Unmarshal(bytes.NewReader(raw), &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// QueryTotalSupply is a utility for querying the USDN total supply.
func QueryTotalSupply(ctx context.Context, validator *cosmos.ChainNode) (math.Int, error) {
	raw, _, err := validator.ExecQuery(ctx, "bank", "total-supply-of", "uusdn")
	if err != nil {
		return math.ZeroInt(), err
	}

	var res banktypes.QuerySupplyOfResponse
	err = jsonpb.Unmarshal(bytes.NewReader(raw), &res)
	if err != nil {
		return math.ZeroInt(), err
	}

	return res.Amount.Amount, nil
}
