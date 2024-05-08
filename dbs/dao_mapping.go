// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package dbs

import (
	"github.com/iwind/TeaGo/logs"
	"reflect"
	"sync"
)

var daoMapping = map[string]DAOInterface{}
var daoMappingLocker = &sync.RWMutex{}

// NewDAO create new dao
func NewDAO(daoPointer any) any {
	if daoPointer == nil {
		logs.Println("[DB]NewDAO(nil): 'daoPointer' should not be nil")
		return nil
	}

	daoInterface, ok := daoPointer.(DAOInterface)
	if !ok {
		logs.Println("[DB]NewDAO(): 'daoPointer' should implement methods in DAOInterface")
		return nil
	}
	if daoInitBeforeCallback != nil {
		daoInitBeforeCallback(daoInterface)
	}

	var pointerType = reflect.TypeOf(daoPointer).String()

	daoMappingLocker.RLock()

	// find the dao in mapping
	cachedDAO, found := daoMapping[pointerType]

	daoMappingLocker.RUnlock()

	if found {
		return cachedDAO
	}

	daoMappingLocker.Lock()
	defer daoMappingLocker.Unlock()

	// check again
	cachedDAO, found = daoMapping[pointerType]
	if found {
		return cachedDAO
	}

	// call Init() method in DAO
	var isInitialized bool
	{
		initCaller, isCaller := daoPointer.(interface {
			Init()
		})
		if isCaller {
			initCaller.Init()
			isInitialized = true
		}
	}

	if !isInitialized {
		initCaller, isCaller := daoPointer.(interface {
			Init() error
		})
		if isCaller {
			err := initCaller.Init()
			if err != nil {
				if daoInitErrorCallback != nil {
					err = daoInitErrorCallback(daoInterface, err)
				}
				if err != nil {
					logs.Println("[DB]init '" + pointerType + "' failed: " + err.Error())
				}
			}
		}
	}

	daoMapping[pointerType] = daoInterface
	return daoPointer
}

func AllDAOMapping() map[string]DAOInterface {
	return daoMapping
}
