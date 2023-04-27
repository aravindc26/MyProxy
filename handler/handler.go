package handler

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aravindc26/go-mysql/client"
	"github.com/aravindc26/go-mysql/mysql"
)

type ProxyHandler struct {
	pool *client.Pool
}

func NewProxyHandler(pool *client.Pool) *ProxyHandler {
	return &ProxyHandler{pool: pool}
}

func (ph *ProxyHandler) UseDB(dbName string) error {
	conn, err := ph.pool.GetConn(context.Background())
	if err != nil {
		log.Printf("error fetching connection %v", err)
		return err
	}

	err = conn.UseDB(dbName)
	return err
}

func (ph *ProxyHandler) HandleQuery(query string) (*mysql.Result, error) {
	conn, err := ph.pool.GetConn(context.Background())
	if err != nil {
		log.Printf("error fetching connection %v", err)
		return nil, err
	}

	return conn.Execute(query)
}

func (ph *ProxyHandler) HandleFieldList(table string, fieldWildcard string) ([]*mysql.Field, error) {
	conn, err := ph.pool.GetConn(context.Background())
	if err != nil {
		log.Printf("error fetching connection %v", err)
		return nil, err
	}

	return conn.FieldList(table, fieldWildcard)
}

func (ph *ProxyHandler) HandleStmtPrepare(query string) (params int, columns int, ct interface{}, err error) {
	conn, err := ph.pool.GetConn(context.Background())
	if err != nil {
		log.Printf("error fetching connection %v", err)
		return 0, 0, nil, err
	}

	st, err := conn.Prepare(query)
	if err != nil {
		log.Printf("error preparing statement %v", err)
		return 0, 0, nil, err
	}

	return st.ParamNum(), st.ColumnNum(), st, nil
}

func (ph *ProxyHandler) HandleStmtExecute(ctx interface{}, query string, args []interface{}) (*mysql.Result, error) {
	st, ok := ctx.(*client.Stmt)
	if !ok {
		log.Printf("error casting context to *client.Stmt")
		return nil, errors.New("error casting context to *client.Stmt")
	}

	return st.Execute(args)
}

func (ph *ProxyHandler) HandleStmtClose(ctx interface{}) error {
	st, ok := ctx.(*client.Stmt)
	if !ok {
		log.Printf("error casting context to *client.Stmt")
		return errors.New("error casting context to *client.Stmt")
	}

	return st.Close()
}

func (ph *ProxyHandler) HandleOtherCommand(cmd byte, data []byte) error {
	return fmt.Errorf("error handling cmd %v, %v", cmd, data)
}
