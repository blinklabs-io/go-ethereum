// Copyright 2026 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

func TestTracingStateReadsDoNotAffectBlockAccessList(t *testing.T) {
	statedb, err := state.New(types.EmptyRootHash, state.NewDatabaseForTesting())
	if err != nil {
		t.Fatal(err)
	}
	config := *params.MergedTestChainConfig
	config.AmsterdamTime = new(uint64)
	blockContext := vm.BlockContext{BlockNumber: new(big.Int)}
	rules := config.Rules(blockContext.BlockNumber, true, 0)
	for _, test := range []struct {
		name  string
		state vm.StateDB
	}{
		{name: "plain", state: statedb},
		{name: "hooked", state: state.NewHookedState(statedb, new(tracing.Hooks))},
	} {
		t.Run(test.name, func(t *testing.T) {
			statedb.Prepare(rules, common.Address{}, common.Address{}, nil, nil, nil)
			evm := vm.NewEVM(blockContext, test.state, &config, vm.Config{})
			defer evm.Release()

			tracingState := evm.GetVMContext().StateDB
			addr := common.HexToAddress("0x1234")
			key := common.HexToHash("0x5678")
			tracingState.GetBalance(addr)
			tracingState.GetNonce(addr)
			tracingState.GetCode(addr)
			tracingState.GetCodeHash(addr)
			tracingState.GetState(addr, key)
			tracingState.Exist(addr)

			accessList := statedb.Finalise(rules)
			if len(accessList.Accounts) != 0 {
				t.Fatalf("tracer reads changed block access list: %s", accessList.PrettyPrint())
			}
		})
	}
}
