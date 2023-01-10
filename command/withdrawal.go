package command

import (
	"errors"

	"gihub.com/victorfernandesraton/dev-api-rest/domain"
)

var InsuficientBalanceError = errors.New("insuficient balance to be withdrawal")

type accountRepositoryInWithdrawalCommand interface {
	FindByAccountNumberAndAgency(uint64, uint64) (*domain.Account, error)
	Update(*domain.Account) error
}

type WithdrawalCommand struct {
	AccountRepository accountRepositoryInWithdrawalCommand
}

func (c *WithdrawalCommand) Execute(accountNumber, agency, ammount uint64) (*domain.Account, error) {
	account, err := c.AccountRepository.FindByAccountNumberAndAgency(accountNumber, agency)
	if err != nil {
		return nil, err
	}

	if account.Balance < ammount {
		return nil, InsuficientBalanceError
	}

	account.Balance = account.Balance - ammount

	if err = c.AccountRepository.Update(account); err != nil {
		return nil, err
	}

	return account, nil
}
