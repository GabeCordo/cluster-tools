package cache

import (
	"github.com/GabeCordo/cluster-tools/internal/core/cache/local"
	"github.com/GabeCordo/cluster-tools/internal/core/thread"
	"github.com/GabeCordo/toolchain/logging"
	"testing"
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
	thread, _ := New(cfg, logger, c, irc, in, out, in, out)

	return thread
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

	if thread := GenerateTestCacheThread(c1, c2); thread == nil {
		t.Error("expected success when creating cache thread")
	}
}
