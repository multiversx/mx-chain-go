package elrond_go

import (
	"fmt"
	"testing"
)

type AccountsAdapter interface {
	Prune()
	Rollback()
}

type AccountsDB struct {
}

func (a *AccountsDB) Prune() {
	fmt.Println("AccountsDB prune")
}

func (a *AccountsDB) Rollback() {
	fmt.Println("AccountsDB rollback")
}

//--------------------------------------

type Wrapper1AccountsDB struct {
	*AccountsDB
}

func (a *Wrapper1AccountsDB) Prune() {
	fmt.Println("Wrapper1AccountsDB rollback calling:")
	a.AccountsDB.Prune()
}

func (a *Wrapper1AccountsDB) Rollback() {
	fmt.Println("Wrapper1AccountsDB rollback")
}

//--------------------------------------

type Wrapper2AccountsDB struct {
	*AccountsDB
}

func (a *Wrapper2AccountsDB) Prune() {
	fmt.Println("Wrapper2AccountsDB rollback calling:")
	a.AccountsDB.Prune()
}

func (a *Wrapper2AccountsDB) Rollback() {
	fmt.Println("Wrapper2AccountsDB rollback")
}

func TestA(t *testing.T) {
	w1 := &Wrapper1AccountsDB{}
	w2 := &Wrapper2AccountsDB{}
	a := &AccountsDB{}

	printMe(w1)
	fmt.Println("-------")
	printMe(w2)
	fmt.Println("-------")
	printMe(a)
}

func printMe(a AccountsAdapter) {
	a.Prune()
}
