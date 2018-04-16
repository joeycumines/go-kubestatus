package kubestatus

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"errors"
	"sync"
	"time"
	"context"
	"github.com/google/uuid"
	"github.com/joeycumines/go-detect-cycle/floyds"
	"strings"
	"net/http"
)

type (
	// Service is an instance of the status server
	Service struct {
		config Config

		ctx    context.Context
		cancel context.CancelFunc

		engine *gin.Engine

		init  sync.Once
		mutex sync.Mutex
		fatal FatalError

		uuid    [16]byte
		started time.Time
	}

	// FatalError models an error that occurred within the service, a non-nil Error would indicate that the server is
	// stopped.
	FatalError struct {
		Error   error
		Time    time.Time
		Runtime time.Duration
	}
)

// NewService constructs a new status.Service
func NewService(config Config) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("kubestatus.NewService failed validation for config: %s", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &Service{
		config: config,
		ctx:    ctx,
		cancel: cancel,
		engine: gin.New(),
		uuid:   uuid.New(),
		fatal: FatalError{
			Error: errors.New("kubestatus.Service has not been started yet"),
		},
	}

	service.engine.Use(config.GinHandlers...)

	service.engine.GET(
		"/healthz",
		func(i *gin.Context) {
			status := service.Health()
			i.JSON(status.Code, status)
		},
	)

	service.engine.GET(
		"/readiness",
		func(i *gin.Context) {
			UUIDs := make([]string, 0)
			for _, UUID := range strings.Split(i.Query("uuids"), ",") {
				UUID = strings.TrimSpace(UUID)
				if UUID == "" {
					continue
				}
				UUIDs = append(UUIDs, UUID)
			}
			status := service.Readiness(UUIDs...)
			i.JSON(status.Code, status)
		},
	)

	return service, nil
}

// Validate returns an error if the service wasn't initialised properly
func (s *Service) Validate() error {
	err := func() error {
		if s == nil {
			return errors.New("nil pointer")
		}
		if s.engine == nil {
			return errors.New("nil engine")
		}
		return nil
	}()
	if err == nil {
		return nil
	}
	return fmt.Errorf("kubestatus.Service.Validate failed: %s", err.Error())
}

func (s *Service) ensure() {
	err := s.Validate()
	if err == nil {
		return
	}
	panic(err)
}

// Start initialises the http server, may only happen once, and runs the http server in the background
func (s *Service) Start() error {
	s.ensure()
	err := errors.New("kubestatus.Service.Start may only be called once")
	s.init.Do(func() {
		func() {
			s.mutex.Lock()
			defer s.mutex.Unlock()
			s.started = time.Now()
			s.fatal = FatalError{}
		}()
		go s.start()
		timer := time.NewTimer(s.config.StartWait)
		defer timer.Stop()
		select {
		case <-s.ctx.Done():
		case <-timer.C:
		}
		err = s.Fatal().Error
	})
	if err == nil {
		_, err = Client{Addresses: []string{s.config.URL()}}.Health()
	}
	return err
}

func (s *Service) start() {
	defer s.cancel()
	fatalError := errors.New("unknown error")
	defer func() {
		stopped := time.Now()
		s.mutex.Lock()
		defer s.mutex.Unlock()
		s.fatal = FatalError{
			Error:   fatalError,
			Time:    stopped,
			Runtime: time.Duration(stopped.UnixNano() - s.started.UnixNano()),
		}
	}()
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		fatalError = fmt.Errorf("recovered from panic (%T): %+v", r, r)
	}()
	if err := s.engine.Run(fmt.Sprintf("%s:%d", s.config.Hostname, s.config.Port)); err != nil {
		fatalError = err
	}
}

// Ctx return the service's context, which will cancel once the service has been started then stopped
func (s *Service) Ctx() context.Context {
	s.ensure()
	return s.ctx
}

// Fatal returns any fatal error that occurred within the service
func (s *Service) Fatal() FatalError {
	s.ensure()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.fatal
}

// Health returns the health of the service
func (s *Service) Health() Status {
	s.ensure()
	err := s.Fatal().Error
	if err == nil {
		err = s.config.HealthHandler()
	}
	return NewStatus(s.uuid, s.started, err)
}

// Readiness returns the readiness of the service, taking any number of previous UUIDs (oldest first)
func (s *Service) Readiness(UUIDs ... string) Status {
	s.ensure()

	// test for fatal error
	if err := s.Fatal().Error; err != nil {
		return NewStatus(s.uuid, s.started, err)
	}

	UUIDs = append(UUIDs, uuid.UUID(s.uuid).String())

	// test for circular references
	cycle := floyds.NewBranchingDetector(UUIDs[0], nil)
	for _, UUID := range UUIDs[1:] {
		cycle = cycle.Hare(UUID)
		if !cycle.Ok() {
			status := NewStatus(
				s.uuid,
				s.started,
				fmt.Errorf("cyclic dependency detected for UUID list: %s", strings.Join(UUIDs, ",")),
			)
			status.Code = http.StatusLoopDetected
			return status
		}
	}

	// test the local readiness handler
	if err := s.config.ReadinessHandler(); err != nil {
		return NewStatus(s.uuid, s.started, err)
	}

	// test the remote readiness handler, which passes down the UUID list for circular ref checking
	if _, err := (Client{Addresses: s.config.Dependencies, UUIDs: UUIDs}).Readiness(); err != nil {
		return NewStatus(s.uuid, s.started, err)
	}

	return NewStatus(s.uuid, s.started, nil)
}

// UUID returns this service's UUID
func (s *Service) UUID() [16]byte {
	s.ensure()
	return s.uuid
}
