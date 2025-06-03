package cache

import (
	"testing"

	"github.com/GabeCordo/Flock/internal/core/cache/local"
	"github.com/GabeCordo/Flock/internal/core/thread"
	"github.com/GabeCordo/toolchain/logging"
)

func GenerateTestCacheThread(in chan thread.Request, out chan thread.Response) *Thread {

	var logger *logging.Logger
	if l, err := logging.NewLogger("cache"); err != nil {
	} else {
		logger = l
	}

	cfg := &Config{Debug: true}

	irc := make(chan thread.InterruptEvent, 10)
	c := local.NewCache(100)
	th, _ := New(cfg, logger, c, irc, in, out, in, out)
	return th
}

func TestNewNilArguments(t *testing.T) {

	_, err := New(nil, nil, nil)
	if err == nil {
		t.Error("excepted thread to reject nil arguments for pipeline or logger")
	}
}

func TestNew(t *testing.T) {

	c1 := make(chan thread.Request, 1)
	c2 := make(chan thread.Response, 1)

	if th := GenerateTestCacheThread(c1, c2); th == nil {
		t.Error("expected success when creating cache thread")
	}
}
