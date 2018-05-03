/*
   Copyright 2018 Joseph Cumines

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */

package kubestatus

import (
	"testing"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestNewService(t *testing.T) {
	config := NewConfig()
	config.Port = 9050
	config.GinHandlers = nil
	gin.SetMode(gin.ReleaseMode)
	config.ReadinessHandler = func() error {
		return errors.New("never_ready")
	}
	config.HealthHandler = func() error {
		return nil
	}
	service, err := NewService(config)
	if err != nil {
		t.Fatal(err)
	}
	if service == nil {
		t.Fatal("expected non-nil service")
	}
	err = service.Start()
	if err != nil {
		t.Error(err)
	}
	err = service.Start()
	if err == nil || err.Error() != "kubestatus.Service.Start may only be called once" {
		t.Error("expected only start once", err)
	}

	// check health
	health := service.Health()
	if health.UUID != uuid.UUID(service.UUID()).String() {
		t.Error("bad uuid", health.UUID)
	}
	if health.Success != true || health.Code != 200 || health.Message != "OK" {
		t.Error(health)
	}
	statuses, err := Client{
		Addresses: []string{"http://localhost:9050"},
	}.Health()
	if err != nil || statuses[0] == nil {
		t.Error(statuses, err)
	} else {
		statuses[0].Uptime = health.Uptime
		if *statuses[0] != health {
			t.Error(*statuses[0])
		}
	}

	// check readiness
	readiness := service.Readiness()
	if readiness.UUID != uuid.UUID(service.UUID()).String() {
		t.Error("bad uuid", health.UUID)
	}
	if readiness.Success != false || readiness.Code != 503 || readiness.Message != "never_ready" {
		t.Error(readiness)
	}
	statuses, err = Client{
		Addresses: []string{"http://localhost:9050"},
	}.Readiness()
	if err == nil || statuses[0] == nil {
		t.Error(statuses, err)
	} else {
		statuses[0].Uptime = readiness.Uptime
		if *statuses[0] != readiness {
			t.Error(*statuses[0])
		}
	}

	secondService, err := NewService(config)
	if err != nil || secondService == nil {
		t.Fatal(err)
	}
	err = secondService.Start()
	if err == nil {
		t.Error("expected an error")
	}

	// ensure state
	if secondService.UUID() == service.UUID() {
		t.Error("expected unique uuids")
	}
	if service.Ctx().Err() != nil {
		t.Error("service should not be done")
	}
	if secondService.Ctx().Err() == nil {
		t.Error("secondService should not be done")
	}
	if secondService.Fatal().Error == nil {
		t.Error("expected fatal error")
	}
}

func TestNewService_configUUID(t *testing.T) {
	config := NewConfig()
	config.ReadinessHandler = func() error {
		return errors.New("never_ready")
	}
	config.HealthHandler = func() error {
		return nil
	}

	if config.UUID != [16]byte{} {
		t.Fatal("unexpected config", config.UUID)
	}

	expected := [16]byte{
		1,
		2,
		3,
		4,
		5,
		6,
		7,
		8,
		9,
		10,
		11,
		12,
		13,
		14,
		15,
		16,
	}

	config.UUID = expected

	service, err := NewService(config)

	if err != nil {
		t.Fatal(err)
	}

	if service.uuid != expected {
		t.Fatal("unexpected service", service.uuid)
	}
}
