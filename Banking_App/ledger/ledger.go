package ledger

import (
	"fmt"
)

type AccountBalanceReader interface {
	GetBankTotalBalance(bankID int) (float64, error)
}

type InterbankLedger struct {
	balances  map[int]map[int]float64
	accountDB AccountBalanceReader
}

func NewInterbankLedger(accountDB AccountBalanceReader) *InterbankLedger {
	return &InterbankLedger{
		balances:  make(map[int]map[int]float64),
		accountDB: accountDB,
	}
}

func (il *InterbankLedger) RecordTransfer(fromBankID, toBankID int, amount float64) error {
	if fromBankID == toBankID {
		return fmt.Errorf("invalid transfer: cannot transfer to the same bank (Bank ID: %d)", fromBankID)
	}
	if amount <= 0 {
		return fmt.Errorf("invalid transfer: amount must be positive (Amount: %.2f)", amount)
	}

	remainingAmount := il.settleOppositeBalance(fromBankID, toBankID, amount)

	if remainingAmount > 0 {
		il.addTransferBalance(fromBankID, toBankID, remainingAmount)
	}
	return nil
}

func (il *InterbankLedger) OwedAmount(fromBankID, toBankID int) float64 {
	if _, ok := il.balances[fromBankID]; !ok {
		return 0
	}
	return il.balances[fromBankID][toBankID]
}

func (il *InterbankLedger) AllBalances() map[int]map[int]float64 {
	copyMap := make(map[int]map[int]float64, len(il.balances))
	for from, inner := range il.balances {
		innerCopy := make(map[int]float64, len(inner))
		for to, amt := range inner {
			innerCopy[to] = amt
		}
		copyMap[from] = innerCopy
	}
	return copyMap
}

func (il *InterbankLedger) GetNetBankPosition(bankID int) (actualBalance, totalReceivable, totalOwed float64, err error) {
	totalOwed = il.calculateTotalOwed(bankID)
	totalReceivable = il.calculateTotalReceivable(bankID)

	actualBalance, err = il.accountDB.GetBankTotalBalance(bankID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to retrieve actual bank balance for Bank ID %d: %w", bankID, err)
	}
	return actualBalance, totalReceivable, totalOwed, nil
}

func (il *InterbankLedger) settleOppositeBalance(fromID, toID int, amount float64) float64 {
	oppositeAmt, exists := il.balances[toID][fromID]
	if !exists || oppositeAmt <= 0 {
		return amount
	}

	if amount >= oppositeAmt {
		delete(il.balances[toID], fromID)
		if len(il.balances[toID]) == 0 {
			delete(il.balances, toID)
		}
		return amount - oppositeAmt
	} else {
		il.balances[toID][fromID] -= amount
		return 0
	}
}

func (il *InterbankLedger) addTransferBalance(fromID, toID int, amount float64) {
	if amount <= 0 {
		return
	}
	if il.balances[fromID] == nil {
		il.balances[fromID] = make(map[int]float64)
	}
	il.balances[fromID][toID] += amount
}

func (il *InterbankLedger) calculateTotalOwed(fromID int) float64 {
	var total float64
	if debts, ok := il.balances[fromID]; ok {
		for _, amt := range debts {
			total += amt
		}
	}
	return total
}

func (il *InterbankLedger) calculateTotalReceivable(toID int) float64 {
	var total float64
	for fromID, debts := range il.balances {
		if fromID == toID {
			continue
		}
		total += debts[toID]
	}
	return total
}
