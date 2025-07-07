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

// RecordTransfer processes a transfer of amount between two banks in the ledger
func (il *InterbankLedger) RecordTransfer(fromBankID, toBankID int, amount float64) {
	if fromBankID == toBankID || amount <= 0 {
		return
	}

	opposite := il.balances[toBankID][fromBankID]
	if opposite > 0 {
		if amount >= opposite {
			delete(il.balances[toBankID], fromBankID)
			amount -= opposite
		} else {
			il.balances[toBankID][fromBankID] -= amount
			return
		}
	}

	if amount > 0 {
		if il.balances[fromBankID] == nil {
			il.balances[fromBankID] = make(map[int]float64)
		}
		il.balances[fromBankID][toBankID] += amount
	}
}

// OwedAmount returns how much fromBankID owes to toBankID
func (il *InterbankLedger) OwedAmount(fromBankID, toBankID int) float64 {
	return il.balances[fromBankID][toBankID]
}

// AllBalances returns a deep copy of the ledger's balances
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

// UpdateLedgerFromTransaction is an alias method for RecordTransfer
func (il *InterbankLedger) UpdateLedgerFromTransaction(fromBankID, toBankID int, amount float64) {
	il.RecordTransfer(fromBankID, toBankID, amount)
}

// GetNetBankPosition returns the actual balance, receivables, and owed amounts for a given bank
func (il *InterbankLedger) GetNetBankPosition(bankID int) (actualBalance, receivable, owed float64, err error) {
	for _, toMap := range il.balances[bankID] {
		owed += toMap
	}
	for fromID, toMap := range il.balances {
		if fromID == bankID {
			continue
		}
		receivable += toMap[bankID]
	}
	actualBalance, err = il.accountDB.GetBankTotalBalance(bankID)
	if err != nil {
		return 0, 0, 0, apperror.NewBankError("get balance", err.Error())
	}
	return actualBalance, receivable, owed, nil
}
