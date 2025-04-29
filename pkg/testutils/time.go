package testutils

import (
	"github.com/agiledragon/gomonkey/v2"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type TestTimer struct {
	Timestamp map[string]time.Time
	Patches   gomonkey.Patches
	mu        sync.Mutex
}

func (t *TestTimer) GetTestFunc() string {
	for i := range 100 {
		pc, _, _, _ := runtime.Caller(i)
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			break
		}
		name := fn.Name()
		re := regexp.MustCompile(`^.*\.`)
		name = re.ReplaceAllString(name, "")

		if strings.HasPrefix(name, "Test") {
			return name
		}
	}
	return "main"
}

func (t *TestTimer) Now() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Timestamp[t.GetTestFunc()]
}

func (t *TestTimer) Sleep(delay time.Duration) {
	t.mu.Lock()
	f := t.GetTestFunc()
	t.Timestamp[f] = t.Timestamp[f].Add(delay)
	t.mu.Unlock()
}

func (t *TestTimer) Patch() {
	patches := gomonkey.ApplyFunc(time.Now, t.Now)
	patches.ApplyFunc(time.Sleep, t.Sleep)
}

var testTimer TestTimer

// PatchTimeFuncs can be used in tests to fake real time.
// It monkey-patches the time.Now and time.Sleep functions and keeps track
// of the current timestamp per test function that initiated the calls.
// It needs to be a singleton as tests can run in parallel.
func PatchTimeFuncs(t *testing.T) {
	testTimer.mu.Lock()
	if testTimer.Timestamp == nil {
		testTimer.Timestamp = map[string]time.Time{}
	}
	testTimer.Timestamp[testTimer.GetTestFunc()] = time.Unix(1745900000, 0)
	testTimer.mu.Unlock()
	testTimer.Patch()
	t.Cleanup(testTimer.Patches.Reset)
}
