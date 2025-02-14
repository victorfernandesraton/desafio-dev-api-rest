package storage

import (
	"context"
	"gihub.com/victorfernandesraton/dev-api-rest/domain"
	"github.com/jackc/pgx/v5"
)

type AccountRepository struct {
	DB *pgx.Conn
}

func (r *AccountRepository) Save(account *domain.Account) error {
	_, err := r.DB.Exec(context.Background(), "insert into public.account(id, cpf, carrier_id, agency, account_number) values ($1, $2, $3, $4, $5)",
		account.ID, account.CPF, account.CarrierId, account.Agency, account.AccountNumber,
	)
	return err
}

func (r *AccountRepository) FindByAccountNumberAndAgency(account uint64, agency uint64) (*domain.Account, error) {
	var result *domain.Account
	rows, err := r.DB.Query(context.Background(), "select id, cpf, carrier_id, balance, status, agency, account_number from public.account a where a.agency = $1 and a.account_number = $2",
		agency, account)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, cpf, carrierId string
		var balance, acountNumber, agencyResult uint64
		var status domain.AccountStatus
		err := rows.Scan(&id, &cpf, &carrierId, &balance, &status, &agencyResult, &acountNumber)
		if err != nil {
			return nil, err
		}
		result = &domain.Account{
			ID:            id,
			CPF:           cpf,
			CarrierId:     carrierId,
			Balance:       balance,
			Status:        status,
			Agency:        agencyResult,
			AccountNumber: acountNumber,
		}
	}

	return result, nil
}

func (r *AccountRepository) GenerateIdForAgency(agency uint64) (uint64, error) {
	var result uint64
	rows, err := r.DB.Query(context.Background(), "select COALESCE(MAX(a.account_number),0) from public.account a where a.agency = $1", agency)
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		err := rows.Scan(&result)
		if err != nil {
			return 0, err
		}
	}

	return result, nil
}

func (r *AccountRepository) UpdateBalance(id string, balance uint64) error {
	_, err := r.DB.Exec(context.Background(),
		`update account
			set balance = $2
		where id = $1`,
		id,
		balance,
	)
	return err
}

func (r *AccountRepository) UpdateStatus(id string, status domain.AccountStatus) error {
	_, err := r.DB.Exec(context.Background(),
		`update account
			set status = $2
		where id = $1`,
		id,
		status,
	)
	return err
}

func (r *AccountRepository) UpdateBalanceTransaction(to, from *domain.Account) error {
	ctx := context.Background()
	tx, err := r.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	for _, v := range []*domain.Account{
		to,
		from,
	} {
		if _, err := tx.Exec(ctx, `update account
			set balance = $2
		where id = $1`, v.ID, v.Balance); err != nil {
			defer tx.Rollback(ctx)
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
