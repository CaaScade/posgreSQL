package executor

import (
	"sync"
	"time"
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

	ObtainToken(string)
	ReturnToken(string)

	SetErrorState(string, error)
}

type Token struct {
	Name   string
	Uuid   string
	Actors map[string]bool
}

var (
	errChan chan error

	doneChan    chan bool
	task        Token
	refreshMap  map[string]chan bool
	tokenReturn chan bool
	refreshLock sync.Mutex
	taskLock    sync.Mutex
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
			taskLock.Lock()
			task.Name = tokens[i].Name
			task.Uuid = tokens[i].Uuid
			taskLock.Unlock()
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

func ObtainToken(actor, uuid string) {
	for {
		taskUuid := task.Uuid
		if taskUuid == uuid {
			addActor(actor)
			return
		}
		<-time.After(1 * time.Second)
	}
}

func ReturnToken(actor string, uuid string) {
	taskUuid := task.Uuid
	taskName := task.Name
	if taskName == actor {
		if taskUuid == uuid {
			drainActor(actor)
		}
	}
}

func addActor(actor string) {
	taskLock.Lock()
	if task.Actors == nil {
		task.Actors = map[string]bool{}
	}
	task.Actors[actor] = true
	taskLock.Unlock()
}

func drainActor(actor string) {
	taskLock.Lock()
	delete(task.Actors, actor)
	if len(task.Actors) == 0 {
		tokenReturn <- true
	}
	taskLock.Unlock()
}

func SetErrorState(uuid string, err error) {
	taskUuid := task.Uuid
	if taskUuid == uuid {
		errChan <- err
	}
}
