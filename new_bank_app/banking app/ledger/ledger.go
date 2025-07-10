package ledger

import (
	"fmt"
)

type Ledger struct {
	balances            map[int]map[int]float64
	getBankTotalBalance func(bankID int) (float64, error)
}

func NewLedger(getBalanceFunc func(bankID int) (float64, error)) *Ledger {
	return &Ledger{
		balances:            make(map[int]map[int]float64),
		getBankTotalBalance: getBalanceFunc,
	}
}

func (l *Ledger) RecordTransfer(fromBankID, toBankID int, amount float64) error {
	if fromBankID == toBankID {
		return fmt.Errorf("invalid transfer: cannot transfer to the same bank (Bank ID: %d)", fromBankID)
	}
	if amount <= 0 {
		return fmt.Errorf("invalid transfer: amount must be positive (Amount: %.2f)", amount)
	}

	remainingAmount := l.settleOppositeBalance(fromBankID, toBankID, amount)

	if remainingAmount > 0 {
		l.addTransferBalance(fromBankID, toBankID, remainingAmount)
	}
	return nil
}

func (l *Ledger) OwedAmount(fromBankID, toBankID int) float64 {
	if _, ok := l.balances[fromBankID]; !ok {
		return 0
	}
	return l.balances[fromBankID][toBankID]
}

func (l *Ledger) AllBalances() map[int]map[int]float64 {
	copyMap := make(map[int]map[int]float64, len(l.balances))
	for from, inner := range l.balances {
		innerCopy := make(map[int]float64, len(inner))
		for to, amt := range inner {
			innerCopy[to] = amt
		}
		copyMap[from] = innerCopy
	}
	return copyMap
}

func (l *Ledger) GetNetBankPosition(bankID int) (actualBalance, totalReceivable, totalOwed float64, err error) {
	totalOwed = l.calculateTotalOwed(bankID)
	totalReceivable = l.calculateTotalReceivable(bankID)

	actualBalance, err = l.getBankTotalBalance(bankID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to retrieve actual bank balance for Bank ID %d: %w", bankID, err)
	}
	return actualBalance, totalReceivable, totalOwed, nil
}

func (l *Ledger) settleOppositeBalance(fromID, toID int, amount float64) float64 {
	oppositeAmt, exists := l.balances[toID][fromID]
	if !exists || oppositeAmt <= 0 {
		return amount
	}

	if amount >= oppositeAmt {
		delete(l.balances[toID], fromID)
		if len(l.balances[toID]) == 0 {
			delete(l.balances, toID)
		}
		return amount - oppositeAmt
	} else {
		l.balances[toID][fromID] -= amount
		return 0
	}
}

func (l *Ledger) addTransferBalance(fromID, toID int, amount float64) {
	if amount <= 0 {
		return
	}
	if l.balances[fromID] == nil {
		l.balances[fromID] = make(map[int]float64)
	}
	l.balances[fromID][toID] += amount
}

func (l *Ledger) calculateTotalOwed(fromID int) float64 {
	var total float64
	if debts, ok := l.balances[fromID]; ok {
		for _, amt := range debts {
			total += amt
		}
	}
	return total
}

func (l *Ledger) calculateTotalReceivable(toID int) float64 {
	var total float64
	for fromID, debts := range l.balances {
		if fromID == toID {
			continue
		}
		total += debts[toID]
	}
	return total
}
