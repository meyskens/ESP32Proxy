package endpoints

import (
	"errors"
	"sync"
)

// EndpointDB contains a set of endpoints
type EndpointDB struct {
	endpointsMutex sync.Mutex
	endpoints      map[string]*EndpointDialer
}

// NewEndpointsDB gives an EndpointDB instance
func NewEndpointsDB() *EndpointDB {
	return &EndpointDB{
		endpoints: map[string]*EndpointDialer{},
	}
}

func (e *EndpointDB) Add(endpoint string, target *EndpointDialer) {
	e.endpointsMutex.Lock()
	defer e.endpointsMutex.Unlock()

	e.endpoints[endpoint] = target
}

func (e *EndpointDB) Remove(endpoint string) {
	e.endpointsMutex.Lock()
	defer e.endpointsMutex.Unlock()

	delete(e.endpoints, endpoint)
}

func (e *EndpointDB) Get(endpoint string) (*EndpointDialer, error) {
	e.endpointsMutex.Lock()
	defer e.endpointsMutex.Unlock()

	ep, ok := e.endpoints[endpoint]
	if !ok {
		return nil, errors.New("endpoint not found")
	}

	return ep, nil
}
