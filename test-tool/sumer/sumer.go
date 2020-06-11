package sumer

import "sync"

var Smap sync.Map
var lmap sync.Map

type Row []uint64

func AddSuccess(id string) {
	ac,_:=lmap.LoadOrStore(id,&sync.Mutex{})
	mux := ac.(*sync.Mutex)
	mux.Lock()
	defer mux.Unlock()

	acs,_:= Smap.LoadOrStore(id,  Row{0,0,0,0,0})
	s:=acs.(Row)
	s[0]=s[0]+1

	Smap.Store(id,s)
}

func AddError(id string,i int) {
	ac,_:=lmap.LoadOrStore(id,&sync.Mutex{})
	mux := ac.(*sync.Mutex)
	mux.Lock()
	defer mux.Unlock()

	acs,_:= Smap.LoadOrStore(id, Row{0,0,0,0,0})
	s:=acs.(Row)
	s[1]=s[i]+1

	Smap.Store(id,s)
}


