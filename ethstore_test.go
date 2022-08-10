package ethstore

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

func TestEthstore(t *testing.T) {
	// in this scenario we want to calculate the eth.store-apr for day 10
	// for simplicity there are only 33 validators with indices 0 to 32
	// all validators have the same start-, end- and effective-balance (respectively: 32 Eth, 32.0032 Eth and 32 Eth)
	// every validator proposed the same amount of blocks during day 10 (besides validator 0 who did not propose a block during day 10)
	// all blocks have the same txFeeSum of 10000 Gwei
	// validator 0 exited on the last epoch of day 9
	// validator 1 exited on the last epoch of day 10
	// validator 2 activated on the second epoch of day 10
	// validator 3 activated on the last epoch of day 10 and deposited 32 Eth extra during day 10
	// validator 4 deposited 100 Eth extra during day 10
	// therefore only 29 validators (indices 4 to 32) should be considered when calculating the eth.store, which is: 365 * (sumOfEndBalances - sumOfStartBalances - sumOfExtraDeposits + sumOfTxFees) / sumOfEffbalancesAtStart
	// given our scenario this should result in 365 * (28*32.0032e18+1*32.0032e18+32e18 - 29*32e18 - 1*32e18 + 10000e9*32*225*29/32) / (32e18*29) = 0.0621640625
	// explaining the numbers:
	// - 365 is the number of days in a year (we ignore leap-years for apr-calculation of eth.store)
	// - 28*32.0032e18+1*32.0032e18+32e18 = sumOfEndBalances = 28 validators each with a balance of 32.0032 eth at the end of the day and one validator deposited extra 100 eth - which is added to the endBalance of the validator
	// - 29*32e18 = sumOfStartBalances = 29 validators each with 32 eth start balance
	// - 1*32e18 = sumOfExtraDeposits = 1 validator deposited 100 eth extra in the set of validators that is considered for the calculation, note that the other deposit should not be considered
	// - 10000e9*32*225*29/32 = sumOfTxFees = 10000 Gwei tx-fee for txs in 32*225 blocks (32 blocks in 225 epochs), but only 29 of the 32 validators who actually propose blocks are in the eth.store validator-set
	// - 32e18*29 = sumOfEffectiveBalances = 29 validators have each an effective balance of 32 eth at the start of the eth.store-day
	// - 0.0621640625 = eth.store-apr = according to the eth.store-calculation validators will earn 6.22% interest in a year

	mocks := map[string]string{
		"/eth/v1/beacon/genesis":          `{"data":{"genesis_time":"1606824023","genesis_validators_root":"0x4b363db94e286120d76eb905340fdd4e54bfe9f06bf33ff6cf5ad27f511bfe95","genesis_fork_version":"0x00000000"}}`,
		"/eth/v1/config/spec":             `{"data":{"CONFIG_NAME":"mainnet","PRESET_BASE":"mainnet","TERMINAL_TOTAL_DIFFICULTY":"115792089237316195423570985008687907853269984665640564039457584007913129638912","TERMINAL_BLOCK_HASH":"0x0000000000000000000000000000000000000000000000000000000000000000","TERMINAL_BLOCK_HASH_ACTIVATION_EPOCH":"18446744073709551615","SAFE_SLOTS_TO_IMPORT_OPTIMISTICALLY":"128","MIN_GENESIS_ACTIVE_VALIDATOR_COUNT":"16384","MIN_GENESIS_TIME":"1606824000","GENESIS_FORK_VERSION":"0x00000000","GENESIS_DELAY":"604800","ALTAIR_FORK_VERSION":"0x01000000","ALTAIR_FORK_EPOCH":"74240","BELLATRIX_FORK_VERSION":"0x02000000","BELLATRIX_FORK_EPOCH":"18446744073709551615","SECONDS_PER_SLOT":"12","SECONDS_PER_ETH1_BLOCK":"14","MIN_VALIDATOR_WITHDRAWABILITY_DELAY":"256","SHARD_COMMITTEE_PERIOD":"256","ETH1_FOLLOW_DISTANCE":"2048","INACTIVITY_SCORE_BIAS":"4","INACTIVITY_SCORE_RECOVERY_RATE":"16","EJECTION_BALANCE":"16000000000","MIN_PER_EPOCH_CHURN_LIMIT":"4","CHURN_LIMIT_QUOTIENT":"65536","PROPOSER_SCORE_BOOST":"40","DEPOSIT_CHAIN_ID":"1","DEPOSIT_NETWORK_ID":"1","DEPOSIT_CONTRACT_ADDRESS":"0x00000000219ab540356cbb839cbe05303d7705fa","MAX_COMMITTEES_PER_SLOT":"64","TARGET_COMMITTEE_SIZE":"128","MAX_VALIDATORS_PER_COMMITTEE":"2048","SHUFFLE_ROUND_COUNT":"90","HYSTERESIS_QUOTIENT":"4","HYSTERESIS_DOWNWARD_MULTIPLIER":"1","HYSTERESIS_UPWARD_MULTIPLIER":"5","SAFE_SLOTS_TO_UPDATE_JUSTIFIED":"8","MIN_DEPOSIT_AMOUNT":"1000000000","MAX_EFFECTIVE_BALANCE":"32000000000","EFFECTIVE_BALANCE_INCREMENT":"1000000000","MIN_ATTESTATION_INCLUSION_DELAY":"1","SLOTS_PER_EPOCH":"32","MIN_SEED_LOOKAHEAD":"1","MAX_SEED_LOOKAHEAD":"4","EPOCHS_PER_ETH1_VOTING_PERIOD":"64","SLOTS_PER_HISTORICAL_ROOT":"8192","MIN_EPOCHS_TO_INACTIVITY_PENALTY":"4","EPOCHS_PER_HISTORICAL_VECTOR":"65536","EPOCHS_PER_SLASHINGS_VECTOR":"8192","HISTORICAL_ROOTS_LIMIT":"16777216","VALIDATOR_REGISTRY_LIMIT":"1099511627776","BASE_REWARD_FACTOR":"64","WHISTLEBLOWER_REWARD_QUOTIENT":"512","PROPOSER_REWARD_QUOTIENT":"8","INACTIVITY_PENALTY_QUOTIENT":"67108864","MIN_SLASHING_PENALTY_QUOTIENT":"128","PROPORTIONAL_SLASHING_MULTIPLIER":"1","MAX_PROPOSER_SLASHINGS":"16","MAX_ATTESTER_SLASHINGS":"2","MAX_ATTESTATIONS":"128","MAX_DEPOSITS":"16","MAX_VOLUNTARY_EXITS":"16","INACTIVITY_PENALTY_QUOTIENT_ALTAIR":"50331648","MIN_SLASHING_PENALTY_QUOTIENT_ALTAIR":"64","PROPORTIONAL_SLASHING_MULTIPLIER_ALTAIR":"2","SYNC_COMMITTEE_SIZE":"512","EPOCHS_PER_SYNC_COMMITTEE_PERIOD":"256","MIN_SYNC_COMMITTEE_PARTICIPANTS":"1","RANDOM_SUBNETS_PER_VALIDATOR":"1","EPOCHS_PER_RANDOM_SUBNET_SUBSCRIPTION":"256","DOMAIN_DEPOSIT":"0x03000000","DOMAIN_SELECTION_PROOF":"0x05000000","DOMAIN_BEACON_ATTESTER":"0x01000000","BLS_WITHDRAWAL_PREFIX":"0x00","TARGET_AGGREGATORS_PER_COMMITTEE":"16","DOMAIN_BEACON_PROPOSER":"0x00000000","DOMAIN_VOLUNTARY_EXIT":"0x04000000","DOMAIN_RANDAO":"0x02000000","DOMAIN_AGGREGATE_AND_PROOF":"0x06000000"}}`,
		"/eth/v1/config/deposit_contract": `{"data":{"chain_id":"1","address":"0x00000000219ab540356cbb839cbe05303d7705fa"}}`,
		"/eth/v1/config/fork_schedule":    `{"data":[{"previous_version":"0x00000000","current_version":"0x00000000","epoch":"0"},{"previous_version":"0x00000000","current_version":"0x01000000","epoch":"74240"}]}`,
		"/eth/v1/node/version":            `{"data":{"version":"Lighthouse/v2.3.1-564d7da/x86_64-linux"}}`,
		"/eth/v2/beacon/blocks/0":         `{"version":"phase0","data":{"message":{"slot":"0","proposer_index":"0","parent_root":"0x0000000000000000000000000000000000000000000000000000000000000000","state_root":"0x7e76880eb67bbdc86250aa578958e9d0675e64e714337855204fb5abaaf82c2b","body":{"randao_reveal":"0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","eth1_data":{"deposit_root":"0x0000000000000000000000000000000000000000000000000000000000000000","deposit_count":"0","block_hash":"0x0000000000000000000000000000000000000000000000000000000000000000"},"graffiti":"0x0000000000000000000000000000000000000000000000000000000000000000","proposer_slashings":[],"attester_slashings":[],"attestations":[],"deposits":[],"voluntary_exits":[]}},"signature":"0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"}}`,
	}

	type MockValidator struct {
		Index     string `json:"index"`
		Balance   string `json:"balance"`
		Status    string `json:"status"`
		Validator struct {
			Pubkey                     string `json:"pubkey"`
			WithdrawalCredentials      string `json:"withdrawal_credentials"`
			EffectiveBalance           string `json:"effective_balance"`
			Slashed                    bool   `json:"slashed"`
			ActivationEligibilityEpoch string `json:"activation_eligibility_epoch"`
			ActivationEpoch            string `json:"activation_epoch"`
			ExitEpoch                  string `json:"exit_epoch"`
			WithdrawableEpoch          string `json:"withdrawable_epoch"`
		} `json:"validator"`
	}

	type MockValidatorsResponse struct {
		Data []MockValidator `json:"data"`
	}

	txFeeGweiPerBlock := uint64(10000)
	numValis := 33

	mockStartValidators := MockValidatorsResponse{make([]MockValidator, numValis)}
	for i := 0; i < numValis; i++ {
		v := MockValidator{}
		v.Index = fmt.Sprintf("%d", i)
		v.Balance = "32000000000"
		v.Status = "active_ongoing"
		v.Validator.Pubkey = fmt.Sprintf("%#096x", i)
		v.Validator.WithdrawalCredentials = fmt.Sprintf("%#064x", i)
		v.Validator.EffectiveBalance = "32000000000"
		v.Validator.Slashed = false
		v.Validator.ActivationEligibilityEpoch = "18446744073709551615"
		v.Validator.ActivationEpoch = "0"
		v.Validator.ExitEpoch = "18446744073709551615"
		v.Validator.WithdrawableEpoch = "18446744073709551615"
		mockStartValidators.Data[i] = v
	}

	mockEndValidators := MockValidatorsResponse{make([]MockValidator, numValis)}
	for i := 0; i < numValis; i++ {
		v := MockValidator{}
		v.Index = fmt.Sprintf("%d", i)
		v.Balance = "32003200000"
		v.Status = "active_ongoing"
		v.Validator.Pubkey = fmt.Sprintf("%#096x", i)
		v.Validator.WithdrawalCredentials = fmt.Sprintf("%#064x", i)
		v.Validator.EffectiveBalance = "32000000000"
		v.Validator.Slashed = false
		v.Validator.ActivationEligibilityEpoch = "18446744073709551615"
		v.Validator.ActivationEpoch = "0"
		v.Validator.ExitEpoch = "18446744073709551615"
		v.Validator.WithdrawableEpoch = "18446744073709551615"
		mockEndValidators.Data[i] = v
	}

	// validator 0 exited on the last epoch of day 9
	mockStartValidators.Data[0].Validator.ExitEpoch = fmt.Sprintf("%d", 10*225-1)
	mockStartValidators.Data[0].Status = "exited_unslashed"
	mockEndValidators.Data[0].Validator.ExitEpoch = fmt.Sprintf("%d", 10*225-1)
	mockEndValidators.Data[0].Status = "exited_unslashed"
	mockEndValidators.Data[0].Balance = "32000000000"

	// validator 1 exited on the last epoch of day 10
	mockEndValidators.Data[1].Validator.ExitEpoch = fmt.Sprintf("%d", 11*225-1)
	mockEndValidators.Data[1].Status = "exited_unslashed"

	// validator 2 activated on the second epoch of day 10
	mockStartValidators.Data[2].Validator.ActivationEpoch = fmt.Sprintf("%d", 10*225+1)
	mockStartValidators.Data[2].Status = "pending_queued"
	mockEndValidators.Data[2].Validator.ActivationEpoch = fmt.Sprintf("%d", 10*225+1)
	mockEndValidators.Data[2].Status = "active_ongoing"

	// validator 3 activated on the last epoch of day 10 and deposited 100 Eth extra during day 10
	mockStartValidators.Data[3].Validator.ActivationEpoch = fmt.Sprintf("%d", 11*225-1)
	mockStartValidators.Data[3].Status = "pending_queued"

	// validator 4 deposited 100 Eth extra during day 10
	mockEndValidators.Data[4].Balance = "64003200000"

	mockStartValidatorsJson, err := json.Marshal(&mockStartValidators)
	if err != nil {
		t.Error(err)
	}

	mockEndValidatorsJson, err := json.Marshal(&mockEndValidators)
	if err != nil {
		t.Error(err)
	}

	mocks["/eth/v1/beacon/states/72000/validators"] = string(mockStartValidatorsJson)
	mocks["/eth/v1/beacon/states/79200/validators"] = string(mockEndValidatorsJson)

	validator4DidExtraDeposit := false
	for i := 10 * 225 * 32; i < 11*225*32; i++ {
		proposer := i%(numValis-1) + 1 // validator with index 0 does not propose blocks on this day
		deposits := "[]"
		if proposer == 4 && !validator4DidExtraDeposit {
			// validator 4 deposited 100 Eth extra during day 10
			validator4DidExtraDeposit = true
			deposits = fmt.Sprintf(`[{ "proof": [ "0xc6efd763704a8c9027488efcf600db43d9c4fad01c2849c34aae7b588bfb108f", "0x928c16bac1799a6cfa95f8112d11b55106ec795c814d13c522b32f27268a45e9", "0x1169c8e967d4cc1c679208635446faaa917c054dcfbdd201a78be9c1756091ec", "0x9e3eeadd60fd8275d89c58aa023d674d4b1579495c3e1b9a7da6287b1d1e21c4", "0x7c082762514951b91903ff93d28724649e9411e7ce44442ffa98658abd2934b3", "0x71a0309da78ef135045aef77a0d4d3a2d8632e9be8c1ff3fbea02b898050f3f7", "0xd88ddfeed400a8755596b21942c1497e114c302e6118290f91e6772976041fa1", "0x87eb0ddba57e35f6d286673802a4af5975e22506c7cf4c64bb6be5ee11527f2c", "0x26846476fd5fc54a5d43385167c95144f2643f533cc85bb9d16b782f8d7db193", "0x506d86582d252405b840018792cad2bf1259f1ef5aa5f887e13cb2f0094f51e1", "0xaf31894d26d5e9cab3824db11ea430ac40da6354fd9e8488ec4a653cbaa759fa", "0x6cf04127db05441cd833107a52be852868890e4317e6a02ab47683aa75964220", "0x63baf863488403a1c9f18f8d752801909cd8f7b1ca65d4e41845d58dce5b6e7e", "0xdf6af5f5bbdb6be9ef8aa618e4bf8073960867171e29676f8b284dea6a08a85e", "0xb58d900f5e182e3c50ef74969ea16c7726c549757cc23523c369587da7293784", "0xd49a7502ffcfb0340b1d7885688500ca308161a7f96b62df9d083b71fcc8f2bb", "0x8fe6b1689256c0d385f42f5bbe2027a22c1996e110ba97c171d3e5948de92beb", "0xaba01bf7fe57be4373f47ff8ea6adc4348fab087b69b2518ce630820f95f4150", "0xd47152335d9460f2b6fb7aba05ced32a52e9f46659ccd3daa2059661d75a6308", "0xf893e908917775b62bff23294dbbe3a1cd8e6cc1c35b4801887b646a6f81f17f", "0xcddba7b592e3133393c16194fac7431abf2f5485ed711db282183c819e08ebaa", "0x8a8d7fe3af8caa085a7639a832001457dfb9128a8061142ad0335629ff23ff9c", "0xfeb3c337d7a51a6fbf00b9e34c52e1c9195c969bd4e7a0bfd51d5c5bed9c1167", "0xe71f0aa83cc32edfbefa9f4d3e0174ca85182eec9f3a09f6a6c0df6377a510d7", "0x31206fa80a50bb6abe29085058f16212212a60eec8f049fecb92d8c8e0a84bc0", "0x21352bfecbeddde993839f614c3dac0a3ee37543f9b412b16199dc158e23b544", "0x619e312724bb6d7c3153ed9de791d764a366b389af13c58bf8a8d90481a46765", "0x7cdd2986268250628d0c10e385c58c6191e6fbe05191bcc04f133f2cea72c1c4", "0x848930bd7ba8cac54661072113fb278869e07bb8587f91392933374d017bcbe1", "0x8869ff2c22b28cc10510d9853292803328be4fb0e80495e8bb8d271f5b889636", "0xb5fe28e79f1b850f8658246ce9b6a1e7b49fc06db7143e8fe0b4f2b0c5523a5c", "0x985e929f70af28d0bdd1a90a808f977f597c7c778c489e98d3bd8910d31ac0f7", "0x3c14060000000000000000000000000000000000000000000000000000000000" ], "data": { "pubkey": "%s", "withdrawal_credentials": "0x00ac3644cd68e81889b3ccc9bb762bc7d5ee98eb6b3abd20e64b3dee06470dd3", "amount": "32000000000", "signature": "0xb1e22bf84074220a072bd1db227dca2736de987aa903e1ced3f13d9f82fd87a6357fb2746ad4fe80d414ccdce98927ff05b5bd3bed2dfe5412e57b413e2af1187c8b7761b78dec9559b41cccebba0daceca4fa4e96eb5c6a7f0dbb64328d88eb" } }]`, mockStartValidators.Data[4].Validator.Pubkey)
		}
		mocks[fmt.Sprintf("/eth/v2/beacon/blocks/%d", i)] = fmt.Sprintf(`{"version":"bellatrix","data":{"message":{"slot":"%d","proposer_index":"%d","parent_root":"0xae77f6e0db57769b5ec6c16c4ef7489ddd47728d98297833b5a1692afc5072cb","state_root":"0x3c900df8e277bade69a1c29a93f9442940fc5e43a96c60dfc33d0f0a54a73af6","body":{"randao_reveal":"0x886b31ed2d6caead1e6632dcaec7edb113789f81dbc101160f903ad72c01429203c15ae75e00bd6987ca5ec79750f9c6040a7805284b24f5b3fa8131579c743e592033de069345ccb4b9a99fd73712d8b2276791847282dbfb7634fcb050ae80","eth1_data":{"deposit_root":"0x9df92d765b5aa041fd4bbe8d5878eb89290efa78e444c1a603eecfae2ea05fa4","deposit_count":"403","block_hash":"0x4d0d1732d9a72d2127ab2ad120e66da738cab3369239ec9debd7aea3b89f9812"},"graffiti":"0x0000000000000000000000000000000000000000000000000000000000000000","proposer_slashings":[],"attester_slashings":[],"attestations":[{"aggregation_bits":"0xf7fa6fffbcbbbf6f","data":{"slot":"357843","index":"0","beacon_block_root":"0xae77f6e0db57769b5ec6c16c4ef7489ddd47728d98297833b5a1692afc5072cb","source":{"epoch":"11181","root":"0xa0d0f93cc58e7e0a6b08c600d2a8054dc41fbadd8aba116e6e8cb1a1870321d0"},"target":{"epoch":"11182","root":"0x82cf146d63ea46194fb6ea4e2c99b244aea76cf8c6546ae09a749a0406d78823"}},"signature":"0xad7d675b775c89fb5c1605f1c91bb595e4feb0a2a0440b23aacfbc6d95daa02e761e8ad48a6cf0dd041d65250a97bf1200e879212f389173cdb2c5792d977411aa44f62eb79e71447f00f2eb02c3aacb4fdc4e939a5d7d01a2198ccdb758b641"}],"deposits":%s,"voluntary_exits":[],"sync_aggregate":{"sync_committee_bits":"0xf74edf53ffdb7f7f7db76efef7fcfb6eff7ffeffbff7f7fddf3f57f7d7fff1b7b7fb3e7bffffff5afe7fffff7fcb437fdffee3efd6dff76df766ffffd7fffff1","sync_committee_signature":"0x98fef94f6488bcb1d1c47517e28683d280c36cfd3caa37403e40a72b0500de7ce84f234760edc17a2bd1031db194570d17af1eb253d4d117f88b39e30ee0ab7c00db268db8369188600a9665708ddd34701840ca1bc1b3c646641b60eda2019d"},"execution_payload":{"parent_hash":"0xca7e7e7fcf3ef35a569c1647d56b11873664e3972d17c5dc339af901230166d5","fee_recipient":"0x8b0c2c4c8eb078bc6c01f48523764c8942c0c6c4","state_root":"0x65ff6f9be55e066f1ed9f5f899752e174c31793034260389316c0ae897483512","receipts_root":"0x1544df33845496bdab8cb97867ec0c6e060ed6690e54c85ae4cb9cc58ddc00dd","logs_bloom":"0x08000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000200000000000000000000000000004000000000000001002000000000000001000000000000000000000000000000020000000100000000000800000000000000000000000000000000080000000000000000000000000000000000000480000000008000000000000000000000001040000000000000000000000000000000000000000000000000000000000000000000000400000000000000004000000001000000000000000020000000000000000000000000000000000000000000000000000000000000000010","prev_randao":"0x3c3397f7c670538c30a11f6c5733e66af09f9a34ab0ef31b0ffa63314b79099f","block_number":"1663387","gas_limit":"30000000","gas_used":"230800","timestamp":"1660027728","extra_data":"0x","base_fee_per_gas":"7","block_hash":"0x8145108c4ba0bd6507019ee9ef1eaa225daa0fd220bfea44f5e1d3b58c313875","transactions":["%#x"]}}},"signature":"0x8b0c109f0148cd7979bc8101f35e909c8b24e08fbfb0a36491270f2d3889c08b71ab83f59f005eff75272627e569f2d91769524dd5790f918955315534e245ad65423fe45f6fb749d9d4cc593c6f56388eef6c5b123b0f7cb526cbdf7fa053c8"}}`, i, proposer, deposits, createTx(txFeeGweiPerBlock))
	}

	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mock, exists := mocks[r.URL.Path]
			if !exists {
				t.Errorf("mock does not exist for request: %v", r.URL.Path)
			}
			w.Write([]byte(mock))
		}),
	)
	defer s.Close()

	// SetDebugLevel(1)
	day, err := Calculate(context.Background(), s.URL, "10")
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", *day)

	extraDepositsWei := decimal.NewFromInt(32e9).Mul(decimal.NewFromInt(1e9))
	endWei := decimal.NewFromInt(29 * 320032e5).Mul(decimal.NewFromInt(1e9)).Add(extraDepositsWei)
	startWei := decimal.NewFromInt(29 * 32e9).Mul(decimal.NewFromInt(1e9))
	consWei := endWei.Sub(startWei).Sub(extraDepositsWei)
	execWei := decimal.NewFromInt(29 * 10000 * 225).Mul(decimal.NewFromInt(1e9))
	eff := decimal.NewFromInt(29 * 32e9).Mul(decimal.NewFromInt(1e9))
	apr := decimal.NewFromInt(365).Mul(consWei.Add(execWei)).Div(eff)

	if day.Day.String() != "10" {
		t.Errorf("wrong Day: %v != %v", day.Day.String(), 10)
	}
	if !day.Apr.Equal(apr) {
		t.Errorf("wrong Apr: %v != %v", day.Apr, apr)
	}
	if day.Validators.IntPart() != 29 {
		t.Errorf("wrong Validators: %v != %v", day.Validators, 29)
	}
	if day.StartEpoch.IntPart() != 2250 {
		t.Errorf("wrong StartEpoch: %v != %v", day.StartEpoch, 2250)
	}
	if !day.StartBalanceGwei.Equal(startWei.Div(decimal.NewFromInt(1e9))) {
		t.Errorf("wrong StartBalanceGwei: %v != %v", day.StartBalanceGwei, startWei.Div(decimal.NewFromInt(1e9)))
	}
	if !day.EndBalanceGwei.Equal(endWei.Div(decimal.NewFromInt(1e9))) {
		t.Errorf("wrong EndBalanceGwei: %v != %v", day.EndBalanceGwei, endWei.Div(decimal.NewFromInt(1e9)))
	}
	if !day.DepositsSumGwei.Equal(extraDepositsWei.Div(decimal.NewFromInt(1e9))) {
		t.Errorf("wrong DepositsSumGwei: %v != %v", day.DepositsSumGwei, extraDepositsWei.Div(decimal.NewFromInt(1e9)))
	}
	if !day.ConsensusRewardsGwei.Equal(consWei.Div(decimal.NewFromInt(1e9))) {
		t.Errorf("wrong ConsensusRewardsGwei: %v != %v", day.ConsensusRewardsGwei, 92800000)
	}
	if !day.TxFeesSumWei.Equal(execWei) {
		t.Errorf("wrong TxFeesSumWei: %v != %v", day.TxFeesSumWei, execWei)
	}
	if !day.TotalRewardsWei.Equal(consWei.Add(execWei)) {
		t.Errorf("wrong TotalRewardsWei: %v != %v", day.TotalRewardsWei, consWei.Add(execWei))
	}
}

func createTx(feeGwei uint64) []byte {
	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	if err != nil {
		log.Fatal(err)
	}
	value := big.NewInt(1e18)
	gasLimit := uint64(feeGwei)
	toAddress := common.HexToAddress("0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d")
	var data []byte
	tx := types.NewTransaction(1, toAddress, value, gasLimit, new(big.Int).SetInt64(1e9), data)
	chainID := new(big.Int).SetInt64(11155111)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	b, err := signedTx.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}
	return b
}