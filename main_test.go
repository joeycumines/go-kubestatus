package kubestatus

import (
	"context"
	"runtime"
	"path"
	"os"
	"testing"
	"os/exec"
	"time"
	"sync"
	"github.com/google/uuid"
)

func findPackagePath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	return path.Dir(filename)
}

var (
	pkgPath  string
	dataPath string
	buildDir string
)

func init() {
	pkgPath = findPackagePath()
	dataPath = path.Join(pkgPath, "testdata")
	buildDir = path.Join(dataPath, "build")
}

func runDebugServer(ctx context.Context, init string, wg *sync.WaitGroup) {
	os.Remove(buildDir)
	os.Mkdir(buildDir, 0777)
	if err := exec.Command("cp", path.Join(dataPath, "main.go"), path.Join(buildDir, "main.go")).Run(); err != nil {
		panic(err)
	}
	func() {
		f, err := os.OpenFile(path.Join(buildDir, "main.go"), os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if _, err = f.WriteString(init); err != nil {
			panic(err)
		}
	}()
	func() {
		cmd := exec.Command("go", "build")
		cmd.Dir = buildDir
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := exec.CommandContext(ctx, path.Join(buildDir, "build")).Run(); err != nil {
			//panic(err)
		}
	}()
	time.Sleep(time.Millisecond * 100)
}

func TestHealth(t *testing.T) {
	wg := new(sync.WaitGroup)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
	}()
	init := `
func init() {
	config = kubestatus.Config{
		Port: 9050,
		HealthHandler: func() error {
			return nil
		},
		ReadinessHandler: func() error {
			return nil
		},
		Dependencies: []string{
		},
	}
}
`
	runDebugServer(ctx, init, wg)

	client := Client{
		Addresses: []string{
			"http://localhost:9050",
		},
	}

	statuses, err := client.Health()
	if err != nil {
		t.Error(err)
	}
	if nil == statuses[0] {
		t.Fatal("expected a status")
	}
	status := *statuses[0]
	if status.Code != 200 {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Message != "OK" {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Success != true {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Started >= time.Now().UnixNano() {
		t.Errorf("unexpected status: %+v", status)
	}
	if _, err := uuid.Parse(status.UUID); err != nil {
		t.Errorf("unexpected status: %+v", status)
		t.Error(err)
	}
}

func TestReadiness(t *testing.T) {
	wg := new(sync.WaitGroup)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
	}()
	init := `
func init() {
	config = kubestatus.Config{
		Port: 9050,
		HealthHandler: func() error {
			return nil
		},
		ReadinessHandler: func() error {
			return nil
		},
		Dependencies: []string{
		},
	}
}
`
	runDebugServer(ctx, init, wg)

	client := Client{
		Addresses: []string{
			"http://localhost:9050",
		},
	}

	statuses, err := client.Readiness()
	if err != nil {
		t.Error(err)
	}
	if nil == statuses[0] {
		t.Fatal("expected a status")
	}
	status := *statuses[0]
	if status.Code != 200 {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Message != "OK" {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Success != true {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Started >= time.Now().UnixNano() {
		t.Errorf("unexpected status: %+v", status)
	}
	if _, err := uuid.Parse(status.UUID); err != nil {
		t.Errorf("unexpected status: %+v", status)
		t.Error(err)
	}
}

func TestHealth_failure(t *testing.T) {
	wg := new(sync.WaitGroup)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
	}()
	init := `
func init() {
	config = kubestatus.Config{
		Port: 9050,
		HealthHandler: func() error {
			return fmt.Errorf("some_error")
		},
		ReadinessHandler: func() error {
			return nil
		},
		Dependencies: []string{
		},
	}
}
`
	runDebugServer(ctx, init, wg)

	client := Client{
		Addresses: []string{
			"http://localhost:9050",
		},
	}

	statuses, err := client.Health()
	if err == nil || err.Error() != "503 Service Unavailable: some_error" {
		t.Error(err)
	}
	if nil == statuses[0] {
		t.Fatal("expected a status")
	}
	status := *statuses[0]
	if status.Code != 503 {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Message != "some_error" {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Success != false {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Started >= time.Now().UnixNano() {
		t.Errorf("unexpected status: %+v", status)
	}
	if _, err := uuid.Parse(status.UUID); err != nil {
		t.Errorf("unexpected status: %+v", status)
		t.Error(err)
	}
}

func TestReadiness_failure(t *testing.T) {
	wg := new(sync.WaitGroup)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
	}()
	init := `
func init() {
	config = kubestatus.Config{
		Port: 9050,
		HealthHandler: func() error {
			return nil
		},
		ReadinessHandler: func() error {
			return fmt.Errorf("some_error")
		},
		Dependencies: []string{
		},
	}
}
`
	runDebugServer(ctx, init, wg)

	client := Client{
		Addresses: []string{
			"http://localhost:9050",
		},
	}

	statuses, err := client.Readiness()
	if err == nil || err.Error() != "503 Service Unavailable: some_error" {
		t.Error(err)
	}
	if nil == statuses[0] {
		t.Fatal("expected a status")
	}
	status := *statuses[0]
	if status.Code != 503 {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Message != "some_error" {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Success != false {
		t.Errorf("unexpected status: %+v", status)
	}
	if status.Started >= time.Now().UnixNano() {
		t.Errorf("unexpected status: %+v", status)
	}
	if _, err := uuid.Parse(status.UUID); err != nil {
		t.Errorf("unexpected status: %+v", status)
		t.Error(err)
	}
}

func TestReadiness_multi(t *testing.T) {
	wg := new(sync.WaitGroup)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		wg.Wait()
	}()

	// 9050 is dependant on 9051
	init := `
func init() {
	config = kubestatus.Config{
		Port: 9050,
		HealthHandler: func() error {
			return nil
		},
		ReadinessHandler: func() error {
			return nil
		},
		Dependencies: []string{
			"http://localhost:9051",
		},
	}
}
`
	runDebugServer(ctx, init, wg)

	client := Client{
		Addresses: []string{
			"http://localhost:9050",
		},
	}

	checkAvailable := func(available bool) {
		if available {
			statuses, err := client.Readiness()
			if err != nil {
				t.Error(err)
			}
			if nil == statuses[0] {
				t.Fatal("expected a status")
			}
			status := *statuses[0]
			if status.Code != 200 {
				t.Errorf("unexpected status: %+v", status)
			}
			if status.Message != "OK" {
				t.Errorf("unexpected status: %+v", status)
			}
			if status.Success != true {
				t.Errorf("unexpected status: %+v", status)
			}
			if status.Started >= time.Now().UnixNano() {
				t.Errorf("unexpected status: %+v", status)
			}
			if _, err := uuid.Parse(status.UUID); err != nil {
				t.Errorf("unexpected status: %+v", status)
				t.Error(err)
			}
		} else {
			// 9050 should be unavailable
			statuses, err := client.Readiness()
			if err == nil {
				t.Error("expected 9050 to be unavailable")
			}
			if nil == statuses[0] {
				t.Fatal("expected a status")
			}
			status := *statuses[0]
			if status.Code != 503 {
				t.Errorf("unexpected status: %+v", status)
			}
			if status.Message == "OK" {
				t.Errorf("unexpected status: %+v", status)
			}
			if status.Success != false {
				t.Errorf("unexpected status: %+v", status)
			}
			if status.Started >= time.Now().UnixNano() {
				t.Errorf("unexpected status: %+v", status)
			}
			if _, err := uuid.Parse(status.UUID); err != nil {
				t.Errorf("unexpected status: %+v", status)
				t.Error(err)
			}
		}
	}

	// 9050 starts unavailable
	checkAvailable(false)

	// but it should be available if we bring up 9051
	func() {
		wg := new(sync.WaitGroup)
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			wg.Wait()
		}()
		init := `
func init() {
	config = kubestatus.Config{
		Port: 9051,
		HealthHandler: func() error {
			return nil
		},
		ReadinessHandler: func() error {
			return nil
		},
		Dependencies: []string{
		},
	}
}
`
		runDebugServer(ctx, init, wg)

		// TEST
		checkAvailable(true)
	}()

	// back to down
	checkAvailable(false)

	// bring up but in an error state
	func() {
		wg := new(sync.WaitGroup)
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			wg.Wait()
		}()
		init := `
func init() {
	config = kubestatus.Config{
		Port: 9051,
		HealthHandler: func() error {
			return nil
		},
		ReadinessHandler: func() error {
			return fmt.Errorf("TWO")
		},
		Dependencies: []string{
		},
	}
}
`
		runDebugServer(ctx, init, wg)

		// TEST
		checkAvailable(false)
	}()

	// test a chain, all up...
	// 9051, requires 9052
	func() {
		wg := new(sync.WaitGroup)
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			wg.Wait()
		}()
		init := `
func init() {
	config = kubestatus.Config{
		Port: 9051,
		HealthHandler: func() error {
			return nil
		},
		ReadinessHandler: func() error {
			return nil
		},
		Dependencies: []string{
			"http://localhost:9052",
		},
	}
}
`
		runDebugServer(ctx, init, wg)

		// 9052, is up
		func() {
			wg := new(sync.WaitGroup)
			ctx, cancel := context.WithCancel(context.Background())
			defer func() {
				cancel()
				wg.Wait()
			}()
			init := `
func init() {
	config = kubestatus.Config{
		Port: 9052,
		HealthHandler: func() error {
			return nil
		},
		ReadinessHandler: func() error {
			return nil
		},
		Dependencies: []string{
		},
	}
}
`
			runDebugServer(ctx, init, wg)
			// TEST
			checkAvailable(true)
		}()

		// back to down
		checkAvailable(false)
	}()

	// back to down
	checkAvailable(false)

	// circular reference test
	// 9051, requires 9052
	func() {
		wg := new(sync.WaitGroup)
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			wg.Wait()
		}()
		init := `
func init() {
	config = kubestatus.Config{
		Port: 9051,
		HealthHandler: func() error {
			return nil
		},
		ReadinessHandler: func() error {
			return nil
		},
		Dependencies: []string{
			"http://localhost:9052",
		},
	}
}
`
		runDebugServer(ctx, init, wg)

		// 9052, is up, but requires 9050, circular ref
		func() {
			wg := new(sync.WaitGroup)
			ctx, cancel := context.WithCancel(context.Background())
			defer func() {
				cancel()
				wg.Wait()
			}()
			init := `
func init() {
	config = kubestatus.Config{
		Port: 9052,
		HealthHandler: func() error {
			return nil
		},
		ReadinessHandler: func() error {
			return nil
		},
		Dependencies: []string{
			"http://localhost:9050",
		},
	}
}
`
			runDebugServer(ctx, init, wg)
			// TEST
			checkAvailable(false)
		}()

		// back to down
		checkAvailable(false)
	}()
}
