package service

import (
	"context"
	"sync/atomic"
)

type (
	Service interface {
		// Init tries to perform the initial initialization of the service, the logic of the function must make sure
		// that all created connections to remote services are in working order and are pinging. Otherwise, the
		// application will need additional error handling.
		Init(ctx context.Context) error
		// Ping will be called by the service controller at regular intervals, it is important that a response with
		// any error will be regarded as an unrecoverable state of the service and will lead to an emergency stop of
		// the application. If the service is not critical for the application, like a memcached, then try to implement
		// the logic of self-diagnosis and service recovery inside Ping, and return the nil as a response even if the
		// recovery failed.
		Ping(ctx context.Context) error
		// Close will be executed when the service controller receives a stop command. Normally, this happens after the
		// main thread of the application has already finished. That is, no more requests from the outside are expected.
		Close() error
	}

	ServiceKeeper struct {
		Services []Service
		state int32 //for control executing stages
	}
)

const (
	srvStateInit int32 = iota
	srvStateReady
	srvStateRunnig
	srvStateShutdown
	srvStateOff

)

func (s *ServiceKeeper) initAllServices(ctx context.Context) error {
	for i := range s.Services {
		if err := s.Services[i].Init(ctx); err != nil {
			return err
		}
	}
	return nil
}


func (s *ServiceKeeper) checkState(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&s.state, old, new)
}

func (s *ServiceKeeper) Init(ctx context.Context) error  {
	if !s.checkState(srvStateInit, srvStateReady) {
		return ErrWrongState
	}

	return s.initAllServices(ctx)
}