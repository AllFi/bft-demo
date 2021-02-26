package main

import (
	"fmt"
	"strconv"

	"github.com/tendermint/tendermint/abci/example/code"
	"github.com/tendermint/tendermint/abci/types"
)

type Application struct {
	types.BaseApplication

	Index     int
	IsCorrect bool
	State     int
	NewState  *int
}

func NewApplication(isCorrect bool, index int) *Application {
	return &Application{IsCorrect: isCorrect, Index: index}
}

func (app *Application) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	fmt.Println("CheckTx " + strconv.Itoa(app.Index) + " " + strconv.FormatBool(app.IsCorrect))
	if app.IsCorrect {
		_, err := strconv.Atoi(string(req.Tx))
		if err != nil {
			return types.ResponseCheckTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Invalid tx format, got %s", string(req.Tx))}
		}
		return types.ResponseCheckTx{Code: code.CodeTypeOK}
	}
	return types.ResponseCheckTx{
		Code: code.CodeTypeUnknownError,
		Log:  fmt.Sprintf("I am malicious", string(req.Tx))}
}

func (app *Application) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	fmt.Println("DeliverTx " + strconv.Itoa(app.Index) + " " + strconv.FormatBool(app.IsCorrect))
	if app.IsCorrect {
		value, err := strconv.Atoi(string(req.Tx))
		if err != nil {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Invalid tx format, got %s", string(req.Tx))}
		}
		app.NewState = &value
		return types.ResponseDeliverTx{Code: code.CodeTypeOK}
	}
	return types.ResponseDeliverTx{
		Code: code.CodeTypeUnknownError,
		Log:  fmt.Sprintf("I am malicious", string(req.Tx))}
}

func (app *Application) Commit() (resp types.ResponseCommit) {
	if app.NewState == nil {
		return types.ResponseCommit{}
	}

	fmt.Println("Commit " + strconv.Itoa(app.Index) + " " + strconv.FormatBool(app.IsCorrect))
	app.State = *app.NewState
	app.NewState = nil
	return types.ResponseCommit{Data: []byte{}}
}

func (app *Application) Query(reqQuery types.RequestQuery) types.ResponseQuery {
	fmt.Println("Query " + strconv.Itoa(app.Index) + " " + strconv.FormatBool(app.IsCorrect))
	switch reqQuery.Path {
	case "lastValue":
		return types.ResponseQuery{Value: []byte(fmt.Sprintf("%v", app.State))}
	case "change":
		app.IsCorrect = !app.IsCorrect
		return types.ResponseQuery{Value: []byte{}}
	default:
		return types.ResponseQuery{Log: fmt.Sprintf("Invalid query path. Expected lastValue, got %v", reqQuery.Path)}
	}
}
