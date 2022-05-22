package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type (
	Resources interface {
		// Init is executed before transferring control to MainFunc. Should initialize resources and check their
		// minimum health. If an error is returned, MainFunc will not be started.
		Init(context.Context) error
		// Watch is executed in the background, monitors the state of resources.
		// Exiting this procedure will immediately stop the application.
		Watch(context.Context) error
		// Stop signals the Watch procedure to terminate the work
		Stop()
		// Release releases the resources. Executed just before exiting the Application.Run
		Release()
	}

	Application struct {
		// MainFunc will run as the main thread of execution when you execute the Run method.
		// Termination of this function will result in the termination of Run, the error that was passed as a
		// result will be thrown as a result of Run execution.
		//
		// The halt channel controls the runtime of the application, as soon as it closes, you need to gracefully
		// complete all current tasks and exit the MainFunc.	
		MainFunc func(ctx context.Context, halt <-chan struct{}) error
		// Resources is an abstraction that represents the resources needed to execute the main thread.
		// The health of resources directly affects the main thread of execution.
		Resources Resources
		// TerminationTimeout limits the time for the main thread to terminate. On normal shutdown,
		// if MainFunc does not return within the allotted time, the job will terminate with an ErrTermTimeout error.
		TerminationTimeout time.Duration
		// InitializationTimeout limits the time to initialize resources.
		// If the resources are not initialized within the allotted time, the application will not be launched
		InitializationTimeout time.Duration

		appState int32
		err error
		mux sync.Mutex
		halt chan struct{}
		done chan struct{}

	}
)

const (
	appStateInit int32 = iota
	appStateRunning
	appStateHalt
	appStateShutdown
)

func (a *Application) Run() error {
	if a.MainFunc == nil {
		// if this func is not set, then nothing to do
		return ErrMainOmitted
	}

	if a.checkState(appStateInit, appStateRunning) {
		// can't enter here twice
		if err := a.init(); err != nil {
			a.err = err
			a.appState = appStateShutdown
			// resources initialisation isn't done
			return err
		}

		// by means servicesRunning we synchronice resources lifecycle with
		// application lifecycle
		var servicesRunning = make(chan struct{})
		if a.Resources != nil {
			go func ()  {
				defer close(servicesRunning) //this signal about Watch stopped
				defer a.shutdown
			}
		}
	}
}

func (a *Application) init() error  {
	if a.Resources != nil {
		ctx, cancel := context.WithTimeout(context.TODO(), a.InitializationTimeout)
		defer cancel()
		return a.Resources.Init(ctx)
	}
	return nil
}

func (a *Application) checkState(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&a.appState, old, new)
}

// Halt signals the application to terminate the current computational processes and prepare to stop the application
func (a *Application) Halt() {
	if a.checkState(appStateRunning, appStateHalt) {
		close(a.halt)
	}
}

// Shutdown stops the application immediately. At this point all calculations should be completed
func (a *Application) Shutdown()  {
	a.Halt()
}