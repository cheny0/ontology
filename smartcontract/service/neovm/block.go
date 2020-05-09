/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
package neovm

import (
	"fmt"
	"os"

	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	vm "github.com/ontio/ontology/vm/neovm"
	vmtypes "github.com/ontio/ontology/vm/neovm/types"
)

// BlockGetTransactionCount put block's transactions count to vm stack
func BlockGetTransactionCount(service *NeoVmService, engine *vm.Executor) error {
	txHash := service.Tx.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d, txHash:%s\n",
		"BlockGetTransactionCount", service.Height, txHash.ToHexString())
	i, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	if block, ok := i.Data.(*types.Block); ok {
		val := vmtypes.VmValueFromInt64(int64(len(block.Transactions)))
		blockHash := block.Hash()
		fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d, blockhash:%s\n",
			"BlockGetTransactionCount", service.Height, blockHash.ToHexString())
		return engine.EvalStack.Push(val)
	}
	return errors.NewErr("[BlockGetTransactionCount] Wrong type ")
}

// BlockGetTransactions put block's transactions to vm stack
func BlockGetTransactions(service *NeoVmService, engine *vm.Executor) error {
	i, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	if block, ok := i.Data.(*types.Block); ok {
		transactions := block.Transactions
		transactionList := make([]vmtypes.VmValue, 0)

		for _, v := range transactions {
			transactionList = append(transactionList, vmtypes.VmValueFromInteropValue(vmtypes.NewInteropValue(v)))
		}

		return engine.EvalStack.PushAsArray(transactionList)
	}
	return errors.NewErr("[BlockGetTransactions] Wrong type ")
}

// BlockGetTransaction put block's transaction to vm stack
func BlockGetTransaction(service *NeoVmService, engine *vm.Executor) error {
	txHash := service.Tx.Hash()
	fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s\n",
		"BlockGetTransaction", service.Height, txHash.ToHexString())
	i, err := engine.EvalStack.PopAsInteropValue()
	if err != nil {
		return err
	}
	index, err := engine.EvalStack.PopAsInt64()
	if err != nil {
		return err
	}
	if block, ok := i.Data.(*types.Block); ok {
		if index < 0 || int(index) >= len(block.Transactions) {
			return errors.NewErr("[BlockGetTransaction] index out of bounds")
		}
		blockHash := block.Hash()
		fmt.Fprintf(os.Stderr, "serviceName:%s, height:%d,txHash:%s, blockhash:%s\n",
			"BlockGetTransaction", service.Height, txHash.ToHexString(), blockHash.ToHexString())
		return engine.EvalStack.PushAsInteropValue(block.Transactions[index])

	}
	return errors.NewErr("[BlockGetTransaction] Wrong type ")
}
