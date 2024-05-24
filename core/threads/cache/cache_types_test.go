package cache

import (
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"testing"
)

func GenerateTestCacheThread(in chan common.ThreadRequest, out chan common.ThreadResponse) *Thread {

	var logger *logging.Logger
	if l, err := logging.NewLogger("cache"); err != nil {
	} else {
		logger = l
	}

	cfg := &Config{Debug: true}

	irc := make(chan common.InterruptEvent, 10)
	thread, _ := New(cfg, logger, irc, in, out, in, out)

	return thread
}

func TestNewNilArguments(t *testing.T) {

	_, err := New(nil, nil)
	if err == nil {
		t.Error("excepted thread to reject nil arguments for config or logger")
	}
}

func TestNew(t *testing.T) {

	c1 := make(chan common.ThreadRequest, 1)
	c2 := make(chan common.ThreadResponse, 1)

	if thread := GenerateTestCacheThread(c1, c2); thread == nil {
		t.Error("expected success when creating cache thread")
	}
}
