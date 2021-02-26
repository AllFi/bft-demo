package app

import (
	"fmt"
	"strconv"
	"time"

	"github.com/tendermint/tendermint/abci/example/code"
	"github.com/tendermint/tendermint/abci/types"
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

	switch app.Status {
	case Correct:
		value, err := strconv.Atoi(string(req.Tx))
		if err != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Invalid tx format, got %s", string(req.Tx))}
		}

		if value%2 != 0 {
			return types.ResponseCheckTx{
				Code: code.CodeTypeUnknownError,
				Log:  fmt.Sprintf("Invalid transaction: value %% 2 != 0, got %s", string(req.Tx))}
		}
		return types.ResponseCheckTx{Code: code.CodeTypeOK}
	case Malicious:
		value, err := strconv.Atoi(string(req.Tx))
		if err != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Invalid tx format, got %s", string(req.Tx))}
		}

		if value%2 == 0 {
			return types.ResponseCheckTx{
				Code: code.CodeTypeUnknownError,
				Log:  fmt.Sprintf("Invalid transaction: value %% 2 == 0, got %s", string(req.Tx))}
		}
		return types.ResponseCheckTx{Code: code.CodeTypeOK}
	case Inaccessible:
		return types.ResponseCheckTx{}
	default:
		panic("unknown app status")
	}
}

func (app *Application) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	defer app.log("DeliverTx", string(req.Tx))

	switch app.Status {
	case Correct:
		value, err := strconv.Atoi(string(req.Tx))
		if err != nil {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Invalid tx format, got %s", string(req.Tx))}
		}

		if value%2 != 0 {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeUnknownError,
				Log:  fmt.Sprintf("Invalid transaction: value %% 2 != 0, got %s", string(req.Tx))}
		}
		app.NewState = &value
		return types.ResponseDeliverTx{Code: code.CodeTypeOK}
	case Malicious:
		value, err := strconv.Atoi(string(req.Tx))
		if err != nil {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Invalid tx format, got %s", string(req.Tx))}
		}

		if value%2 == 0 {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeUnknownError,
				Log:  fmt.Sprintf("Invalid transaction: value %% 2 != 0, got %s", string(req.Tx))}
		}
		app.NewState = &value
		return types.ResponseDeliverTx{Code: code.CodeTypeOK}
	case Inaccessible:
		return types.ResponseDeliverTx{}
	default:
		panic("unknown app status")
	}
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
