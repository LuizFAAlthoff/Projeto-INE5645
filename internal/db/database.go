package db

import "sync"

var Data = map[string]string{}
var Mutex = sync.RWMutex{}
