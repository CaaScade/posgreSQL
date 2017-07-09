package executor

import (
	"sync"
)

/* This package and all other packages withing this subdirectory
 * follow a new design pattern to separate the mess of having the
 * same pipeline for control and data.
 * for eg.
 *   unless i'm using a library function, you'll see that I avoid
 *   using data, err := func() pattern and also avoid the subsequent
 *   if err != nil pattern
 *
 * Secondly, I've wrapped library functions that follow the above
 * pattern into a function thats return only data, and a separate
 * error channel that handles error
 *
 * Thirdly, I've also designed these packages to be self reliant.
 * Each package is independent, and pulls the information it needs
 * from other packages, rather than being useful only in the context
 * of a calling package. This pattern essentially reverses the
 * direction of calls
 *
 * for eg.
 *  client package pulls all its information from the posgresql
 *  package instead of being initialized and co-ordinated by the
 *  calling package
 *
 * Lastly, I've used an anonymous type to descibe the data and
 * behaviors exposed by a package
 *
 * These patterns allow me to cleave the system very clearly, and
 * reason about its various attributes in a straightforward manner
 *
 * This is known as the knowledgebase-actor pattern
 *
 *   The posgresql package is a central co-ordinator of tasks, and
 *   is the central knowledge base. (these can be further divided
 *   into two packages in some cases). The rest of the packages are
 *   independent actors that act on the knowledge base.
 *
 * This model allows both parallel and serial co-ordination of actors
 */

type _ interface {
	Exec(bool, string) error

	ObtainToken(string) string
	ReturnToken(string)

	SetErrorState(string, error)
}

type Token struct {
	Name   string
	Uuid   string
	Actors map[string]bool
	Lock   sync.Mutex
}

var (
	errChan chan error

	doneChan    chan bool
	task        Token
	refreshMap  map[string]chan bool
	tokenReturn chan bool
	refreshLock sync.Mutex
)

func init() {
	errChan = make(chan error, 0)
	doneChan = make(chan bool, 0)
	tokenReturn = make(chan bool, 0)
	refreshMap = map[string]chan bool{}
}

func Exec(tokens []Token) error {
	go func() {
		for i := range tokens {
			<-tokenReturn
			task = tokens[i]
			for i := range refreshMap {
				refreshMap[i] <- true
			}
		}
		<-tokenReturn
		doneChan <- true
	}()

	//start operations
	tokenReturn <- true

	select {
	case err := <-errChan:
		return err
	case <-doneChan:
	}

	return nil
}

func ObtainToken(actor, uuid string) bool {
	tokenRefresh := registerListener(actor)
	for {
		if task.Uuid == uuid {
			addActor(actor)
			unregisterListener(actor)
			return true
		}
		<-tokenRefresh
	}
	// Can never get here
	return false
}

func ReturnToken(actor string, uuid string) {
	if task.Name == actor {
		if task.Uuid == uuid {
			drainActor(actor)
		}
	}
}

func addActor(actor string) {
	task.Lock.Lock()
	task.Actors[actor] = true
	task.Lock.Unlock()
}

func registerListener(actor string) chan bool {
	waitChan := make(chan bool, 0)
	refreshLock.Lock()
	refreshMap[actor] = waitChan
	refreshLock.Unlock()
	return waitChan
}

func unregisterListener(actor string) {
	refreshLock.Lock()
	delete(refreshMap, actor)
	refreshLock.Unlock()
}

func drainActor(actor string) {
	task.Lock.Lock()
	delete(task.Actors, actor)
	task.Lock.Unlock()
	if len(task.Actors) == 0 {
		tokenReturn <- true
	}
}

func SetErrorState(uuid string, err error) {
	if task.Uuid == uuid {
		errChan <- err
	}
}
