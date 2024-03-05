package cache

import (
	"github.com/GabeCordo/cluster-tools/core/threads/common"
	"github.com/GabeCordo/toolchain/logging"
	"testing"
)

func GenerateTestCacheThread(in chan common.CacheRequest, out chan common.CacheResponse) *Thread {

	var logger *logging.Logger
	if l, err := logging.NewLogger("cache"); err != nil {
	} else {
		logger = l
	}

	cfg := &Config{Debug: true}

	irc := make(chan common.InterruptEvent, 10)
	thread, _ := New(cfg, logger, irc, in, out)

	return thread
}

func TestNewNilArguments(t *testing.T) {

	_, err := New(nil, nil)
	if err != nil {
		t.Error("excepted thread to reject nil arguments for config or logger")
	}
}

func TestNew(t *testing.T) {

	var logger *logging.Logger
	if l, err := logging.NewLogger("cache"); err != nil {
		t.Error(err)
		return
	} else {
		logger = l
	}

	cfg := &Config{}

	if _, err := New(cfg, logger); err != nil {
		t.Error(err)
	}
}
