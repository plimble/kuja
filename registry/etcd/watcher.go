package etcd

import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/plimble/kuja/registry"
)

type etcdWatcher struct {
	registry  *EtcdRegistry
	watchChan chan *etcd.Response
	stopChan  chan bool
}

func newEtcdWatcher(r *EtcdRegistry) (*etcdWatcher, error) {
	w := &etcdWatcher{
		registry:  r,
		stopChan:  make(chan bool),
		watchChan: make(chan *etcd.Response),
	}

	go w.registry.client.Watch(r.prefix, 0, true, w.watchChan, w.stopChan)
	go func() {
		for resp := range w.watchChan {
			if resp.Node.Dir {
				continue
			}

			n := decode(resp.Node.Value)
			if n == nil {
				n = decode(resp.PrevNode.Value)
				if n == nil {
					continue
				}
			}
			w.registry.Lock()

			service, ok := w.registry.services[n.Name]
			if !ok {
				if resp.Action == "create" {
					w.registry.services[n.Name] = &registry.Service{
						Name:  n.Name,
						Nodes: []*registry.Node{n},
					}
				}
				w.registry.Unlock()
				continue
			}

			switch resp.Action {
			case "delete":
				for i := 0; i < len(service.Nodes); i++ {
					if n.Id == service.Nodes[i].Id {
						service.Nodes = append(service.Nodes[:i], service.Nodes[i+1:]...)
					}
				}
			case "create":
				service.Nodes = append(service.Nodes, n)
			}
			w.registry.Unlock()
		}
	}()

	return w, nil
}

func (w *etcdWatcher) Stop() {
	w.stopChan <- true
}
