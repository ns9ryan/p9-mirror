package worker

import (
	"log"
	"sync"

	"oa.98ent.com/p9/operator-activity/ent"
	"oa.98ent.com/p9/operator-activity/modules/worker/order"
	"oa.98ent.com/p9/operator-activity/modules/worker/schedule"

	"github.com/zeromicro/go-zero/core/proc"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type WorkerManager struct {
	client *ent.Client
}

func NewWorkerManager(client *ent.Client) *WorkerManager {
	return &WorkerManager{client: client}
}

func (wm *WorkerManager) Init(address, namespace string) {
	c := wm.createTemporalClient(address, namespace)
	defer c.Close()

	registrars := wm.getAllWorkerRegistrars()
	workers := wm.registerAllWorkers(c, registrars)
	wm.startWorkers(workers)
}

func (wm *WorkerManager) getAllWorkerRegistrars() []WorkerRegistrar {
	return []WorkerRegistrar{
		order.NewRegistrar(),
		schedule.NewRegistrar(),
	}
}

func (wm *WorkerManager) createTemporalClient(address, namespace string) client.Client {
	c, err := client.Dial(client.Options{
		HostPort:  address,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalln("Unable to create temporal client", err)
	}
	return c
}

func (wm *WorkerManager) registerAllWorkers(c client.Client, registrars []WorkerRegistrar) []worker.Worker {
	var workers []worker.Worker
	for _, registrar := range registrars {
		w := registrar.Register(c, wm.client)
		log.Printf("registered worker: %s\n", registrar.Name())
		workers = append(workers, w)
	}
	return workers
}

func (wm *WorkerManager) startWorkers(workers []worker.Worker) {
	interruptCh := make(chan interface{})
	go func() {
		<-proc.Done()
		close(interruptCh)
	}()

	var wg sync.WaitGroup
	wg.Add(len(workers))
	for i := range workers {
		w := workers[i]
		go func() {
			defer wg.Done()
			if err := w.Run(interruptCh); err != nil {
				log.Printf("Unable to start worker: %v", err)
			}
		}()
	}
	wg.Wait()
}

func Init(client *ent.Client, address, namespace string) {
	NewWorkerManager(client).Init(address, namespace)
}

type WorkerRegistrar interface {
	Register(c client.Client, entClient *ent.Client) worker.Worker
	Name() string
}
