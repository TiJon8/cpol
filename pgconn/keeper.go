package pgconn

import (
	"context"
	"fmt"

	"github.com/TiJon8/cpol/proto"
)


type SimpleQuery struct {
	// fallback
	pgConn *PgConn
	ctx    context.Context

	resultKeeper *ResultKeeper

	closed bool
	err    error

	quryProcessChan chan struct{}
}

type CommandTag struct {
	String string
}

type ResultKeeper struct {
	commandTag CommandTag

	DataRows [][]proto.DataRow
	RowDescriptions []proto.RowDescription
}


func (sq *SimpleQuery) Close() error {
	for !sq.closed {
		_, err := sq.receiveMessage()
		if err != nil {
			return sq.err
		}
	}
	close(sq.quryProcessChan)
	return sq.err
}

func (sq *SimpleQuery) receiveMessage() (proto.BackendMessage, error) {
	msg, err := sq.pgConn.receiveMessage()
	fmt.Println(msg)
	if err != nil {
		sq.err = err
		sq.closed = true
		sq.pgConn.asyncClose()
		return nil, sq.err
	}

	switch msg.(type) {
	case *proto.ReadyForQuery:
		sq.closed = true
	case *proto.ErrorResponse:
		sq.err = fmt.Errorf("%v", msg)
	}

	return msg, nil
}


func (sq *SimpleQuery) Result() (commandTag CommandTag, err error) {
	for !sq.closed && sq.err == nil {
		msg, err := sq.pgConn.receiveMessage()
		fmt.Println("backend msg", msg)
		if err != nil {
			sq.err = err
			sq.closed = true
			sq.pgConn.asyncClose()
			return CommandTag{}, sq.err
		}
		switch msg := msg.(type) {
		case *proto.RowDescription:
			fmt.Println(msg.Rows)
		case *proto.DataRow:
			fmt.Println(msg.Rows)
		case *proto.CommandComplete:
			rk := CommandTag{String: string(msg.CommandTag)}
			sq.resultKeeper.commandTag = rk
		case *proto.ErrorResponse:
			fmt.Println(msg.Error)
			sq.err = err
			sq.closed = true
		case *proto.ReadyForQuery:
			sq.closed = true
			return sq.resultKeeper.commandTag, err
		}
	}
	return sq.resultKeeper.commandTag, err
}