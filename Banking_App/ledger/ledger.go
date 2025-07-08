package ledger

import (
	"banking-app/apperror"
)

// --- Interfaces required from account system ---

type AccountBalanceReader interface {
	GetBankTotalBalance(bankID int) (float64, error)
}

// --- Ledger Definition ---

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

// --- Public Methods ---

func (il *InterbankLedger) RecordTransfer(fromBankID, toBankID int, amount float64) {
	if !il.isValidTransfer(fromBankID, toBankID, amount) {
		return
	}
	il.settleOppositeBalance(fromBankID, toBankID, &amount)
	il.addTransferBalance(fromBankID, toBankID, amount)
}

func (il *InterbankLedger) OwedAmount(fromBankID, toBankID int) float64 {
	return il.balances[fromBankID][toBankID]
}

func (il *InterbankLedger) AllBalances() map[int]map[int]float64 {
	copy := make(map[int]map[int]float64, len(il.balances))
	for from, inner := range il.balances {
		innerCopy := make(map[int]float64, len(inner))
		for to, amt := range inner {
			innerCopy[to] = amt
		}
		copy[from] = innerCopy
	}
	return copy
}

func (il *InterbankLedger) UpdateLedgerFromTransaction(fromBankID, toBankID int, amount float64) {
	il.RecordTransfer(fromBankID, toBankID, amount)
}

func (il *InterbankLedger) GetNetBankPosition(bankID int) (actualBalance, receivable, owed float64, err error) {
	owed = il.sumOwed(bankID)
	receivable = il.sumReceivable(bankID)

	actualBalance, err = il.accountDB.GetBankTotalBalance(bankID)
	if err != nil {
		return 0, 0, 0, apperror.NewBankError("get balance", err.Error())
	}
	return actualBalance, receivable, owed, nil
}

// --- Internal Helper Methods ---

func (il *InterbankLedger) isValidTransfer(fromID, toID int, amount float64) bool {
	return fromID != toID && amount > 0
}

func (il *InterbankLedger) settleOppositeBalance(fromID, toID int, amount *float64) {
	oppositeAmt := il.balances[toID][fromID]
	if oppositeAmt <= 0 {
		return
	}

	switch {
	case *amount >= oppositeAmt:
		delete(il.balances[toID], fromID)
		*amount -= oppositeAmt
	case *amount < oppositeAmt:
		il.balances[toID][fromID] -= *amount
		*amount = 0
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

func (il *InterbankLedger) sumOwed(fromID int) float64 {
	var total float64
	for _, amt := range il.balances[fromID] {
		total += amt
	}
	return total
}

func (il *InterbankLedger) sumReceivable(toID int) float64 {
	var total float64
	for fromID, toMap := range il.balances {
		if fromID == toID {
			continue
		}
		total += toMap[toID]
	}
	return total
}
