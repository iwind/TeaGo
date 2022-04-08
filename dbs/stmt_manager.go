// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dbs

import (
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"strconv"
	"sync"
	"time"
)

const MaxStmtCount = 4096

func IsPrepareError(err error) bool {
	if err == nil {
		return false
	}
	mysqlErr, isMySQLErr := err.(*mysql.MySQLError)
	return isMySQLErr && mysqlErr.Number == 1461
}

type sqlPreparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

var timestamp = time.Now().Unix()
var timeTicker = time.NewTicker(500 * time.Millisecond)

func init() {
	go func() {
		for range timeTicker.C {
			timestamp = time.Now().Unix()
		}
	}()
}

func unixTime() int64 {
	return timestamp
}

type StmtManager struct {
	stmtMap map[string]*Stmt   // query => *Stmt
	subMap  map[int64][]string // id => [cache keys]
	locker  sync.RWMutex

	isClosed bool
}

func NewStmtManager() *StmtManager {
	var manager = &StmtManager{
		stmtMap: map[string]*Stmt{},
		subMap:  map[int64][]string{},
	}

	return manager
}

// Prepare statement
func (this *StmtManager) Prepare(preparer sqlPreparer, querySQL string) (*Stmt, error) {
	if this.isClosed {
		return nil, errors.New("prepare failed: connection is closed")
	}

	sqlStmt, err := preparer.Prepare(querySQL)
	if err != nil {
		if IsPrepareError(err) {
			// lock for concurrent operation
			this.locker.Lock()
			this.purge()
			this.locker.Unlock()

			// retry
			sqlStmt, err = preparer.Prepare(querySQL)
		}
		if err != nil {
			return nil, err
		}
	}

	return NewStmt(sqlStmt), nil
}

// PrepareOnce prepare statement for reuse
func (this *StmtManager) PrepareOnce(preparer sqlPreparer, querySQL string, parentId int64) (resultStmt *Stmt, wasCached bool, err error) {
	var cacheKey = ""
	if parentId == 0 {
		cacheKey = "0$" + querySQL
	} else {
		cacheKey = strconv.FormatInt(parentId, 10) + "$" + querySQL
	}

	// check if exists
	this.locker.RLock()
	stmt, ok := this.stmtMap[cacheKey]
	if ok {
		stmt.accessAt = timestamp
		this.locker.RUnlock()
		return stmt, true, nil
	}
	this.locker.RUnlock()

	sqlStmt, err := preparer.Prepare(querySQL)
	if err != nil {
		if IsPrepareError(err) {
			// purge once
			this.purge()

			// retry
			sqlStmt, err = preparer.Prepare(querySQL)
		}
		if err != nil {
			return nil, false, err
		}
	}
	stmt = NewStmt(sqlStmt)

	this.locker.Lock()
	defer this.locker.Unlock()

	// exists, check again
	_, exists := this.stmtMap[cacheKey]
	if exists {
		return stmt, false, nil
	}

	// should we purge old statements?
	if len(this.stmtMap) >= MaxStmtCount {
		this.purge()

		// still full
		if len(this.stmtMap) >= MaxStmtCount {
			return stmt, false, nil
		}
	}

	// put stmt into cache map
	this.stmtMap[cacheKey] = stmt
	if parentId > 0 {
		this.subMap[parentId] = append(this.subMap[parentId], cacheKey)
	}

	return stmt, true, nil
}

func (this *StmtManager) Close() error {
	this.isClosed = true

	this.locker.Lock()
	var stmtMap = this.stmtMap
	this.stmtMap = map[string]*Stmt{}
	this.locker.Unlock()

	var firstError error
	for _, stmt := range stmtMap {
		err := stmt.Close()
		if err != nil && firstError == nil {
			firstError = err
		}
	}

	return firstError
}

func (this *StmtManager) CloseId(parentId int64) error {
	// collect dirty stmts
	this.locker.Lock()

	cacheKeys, ok := this.subMap[parentId]
	if !ok {
		this.locker.Unlock()
		return nil
	}
	delete(this.subMap, parentId)

	var dirtyStmts = []*Stmt{}
	for _, cacheKey := range cacheKeys {
		stmt, stmtOk := this.stmtMap[cacheKey]
		if stmtOk {
			dirtyStmts = append(dirtyStmts, stmt)
			delete(this.stmtMap, cacheKey)
		}
	}

	this.locker.Unlock()

	// close dirty stmts
	var firstError error
	for _, stmt := range dirtyStmts {
		err := stmt.Close()
		if err != nil && firstError == nil {
			firstError = err
		}
	}

	return firstError
}

func (this *StmtManager) Len() int {
	this.locker.Lock()
	defer this.locker.Unlock()
	return len(this.stmtMap)
}

func (this *StmtManager) purge() {
	// remove old statements
	var nowTime = time.Now().Unix()
	var total = len(this.stmtMap)
	var count = 0
	for cacheKey, stmt := range this.stmtMap {
		if stmt.accessAt < nowTime-3600 {
			_ = stmt.Close()
			delete(this.stmtMap, cacheKey)
			count++
		}
	}

	// too many left, we purge again
	if count < total/100 {
		for cacheKey, stmt := range this.stmtMap {
			if stmt.accessAt < nowTime-1800 {
				_ = stmt.Close()
				delete(this.stmtMap, cacheKey)
				count++
			}
		}
	}
}
