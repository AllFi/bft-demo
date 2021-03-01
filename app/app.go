package app

import (
	"fmt"
	"strconv"
	"time"

	"github.com/tendermint/tendermint/abci/example/code"
	"github.com/tendermint/tendermint/abci/types"
)

const (
	CodeTxIsNotValid = 5
)

type Application struct {
	types.BaseApplication

	Status   Status
	State    int
	NewState *int
}

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	defer app.log("CheckTx", string(req.Tx))

	value, err := strconv.Atoi(string(req.Tx))
	if err != nil {
		return types.ResponseCheckTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Invalid tx format, tx: %s", string(req.Tx))}
	}

	if (value%2 != 0) == (app.Status == Correct) {
		return types.ResponseCheckTx{
			Code: CodeTxIsNotValid,
			Log:  fmt.Sprintf("Tx is not valid, tx: %s", string(req.Tx))}
	}
	return types.ResponseCheckTx{Code: code.CodeTypeOK}
}

func (app *Application) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	defer app.log("DeliverTx", string(req.Tx))

	checkResp := app.CheckTx(types.RequestCheckTx{Tx: req.Tx})
	if checkResp.Code == code.CodeTypeOK {
		value, _ := strconv.Atoi(string(req.Tx))
		app.NewState = &value
	} else if checkResp.Code == CodeTxIsNotValid {
		app.NewState = &app.State
	}

	return types.ResponseDeliverTx{Code: checkResp.Code, Log: checkResp.Log}
}

func (app *Application) Commit() (resp types.ResponseCommit) {
	if app.NewState == nil {
		return types.ResponseCommit{}
	}

	defer app.log("Commit", "")
	app.State = *app.NewState
	app.NewState = nil
	return types.ResponseCommit{Data: []byte{}}
}

func (app *Application) Query(reqQuery types.RequestQuery) types.ResponseQuery {
	defer app.log("Query", "")

	switch reqQuery.Path {
	case "state":
		return types.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.State))}
	case "status":
		status, err := ParseStatus(string(reqQuery.Data))
		if err != nil {
			return types.ResponseQuery{Code: code.CodeTypeUnknownError}
		}
		app.Status = status
		return types.ResponseQuery{Code: code.CodeTypeOK}
	default:
		return types.ResponseQuery{Log: fmt.Sprintf("Invalid query path. Expected state or status, got %v", reqQuery.Path)}
	}
}

func (app *Application) log(action string, tx string) {
	fmt.Printf("%v Action: %v, Tx: %v, State: %v, Status: %v\n", time.Now().Format("2006-01-02 15:04:05"), action, tx, app.State, app.Status.String())
}
