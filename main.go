package main

import (
	"context"
	"fmt"
	"gihub.com/victorfernandesraton/dev-api-rest/app"
	"gihub.com/victorfernandesraton/dev-api-rest/command"
	"gihub.com/victorfernandesraton/dev-api-rest/infra/event/consume"
	"gihub.com/victorfernandesraton/dev-api-rest/infra/event/emitter"
	"gihub.com/victorfernandesraton/dev-api-rest/infra/event/provider"
	"gihub.com/victorfernandesraton/dev-api-rest/infra/storage"
	"gihub.com/victorfernandesraton/dev-api-rest/query"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	transactionEmitter := emitter.TransactionEmitter{
		Provider: &provider.TransactionProvider{},
	}

	carryRepository := storage.CarrierRepository{
		DB: conn,
	}

	accountRepository := storage.AccountRepository{
		DB: conn,
	}

	transactionRepository := storage.TransactionRepository{
		DB: conn,
	}

	carrierCommand := command.CreateCarrierCommand{
		CaryRepository: &carryRepository,
	}

	createAccountCommand := command.CreateAccountCommand{
		CarrierRepository: &carryRepository,
		AccountRepository: &accountRepository,
	}
	depositCommand := command.DepositCommand{
		AccountRepository: &accountRepository,
	}

	withdrawalAccountCommand := command.WithdrawalCommand{
		AccountRepository:     &accountRepository,
		TransactionRepository: &transactionRepository,
	}
	updateStatusCommand := command.UpdateStatusCommand{
		AccountRepository: &accountRepository,
	}

	transactionCommand := command.TransactionCommand{
		AccountRepository: &accountRepository,
	}

	extractQuery := query.ExtractQuery{
		DB:                conn,
		AccountRepository: &accountRepository,
	}

	app.CarrierControllerFactory(&app.CarrierControllerFactoryParams{
		DefaultControllerFactory: app.DefaultControllerFactory{
			Echo: e,
		},
		CreateCarrierCommand: &carrierCommand,
	})

	app.AccountControllerFactory(&app.AccountControllerFactoryParams{
		DefaultControllerFactory: app.DefaultControllerFactory{
			Echo: e,
		},
		CreateAccountCommand:  &createAccountCommand,
		DepositAccountCommand: &depositCommand,
		WithdrawalCommand:     &withdrawalAccountCommand,
		UpdateStatusCommand:   &updateStatusCommand,
		TransactionEmitter:    &transactionEmitter,
	})

	app.TransactionControllerFactory(&app.TransactionControllerFactoryParams{
		DefaultControllerFactory: app.DefaultControllerFactory{
			Echo: e,
		},
		TransactionCommand: &transactionCommand,
		ExtractQuery:       &extractQuery,
		TransactionEmitter: &transactionEmitter,
	})

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, ":-)")
	})

	queueTransactionConsume := consume.TransactionConsume{
		Provider: &provider.TransactionProvider{},
		DB:       conn,
	}

	go queueTransactionConsume.Listen()

	e.Logger.Fatal(e.Start(":3000"))

}
