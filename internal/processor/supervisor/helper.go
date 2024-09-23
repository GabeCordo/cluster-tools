package supervisor

import (
	"errors"
	"fmt"
	"github.com/GabeCordo/cluster-tools/internal/processor/api"
	"github.com/GabeCordo/toolchain/logging"
)

var CachingInStandaloneErr = errors.New("caching is not supported for standalone mode yet")

type StandardHelper struct {
	host       string
	namespace  string
	pipeline   string
	supervisor uint64
	standalone bool // if we are in standalone mode, don't use the core api
	logger     *logging.Logger
}

func NewHelper(host, namespace, pipeline string, supervisor uint64, standalone bool) *StandardHelper {
	helper := new(StandardHelper)

	// TODO : setup core host link

	helper.host = host
	helper.namespace = namespace
	helper.pipeline = pipeline
	helper.supervisor = supervisor

	if standalone {
		helper.standalone = standalone
		if logger, err := logging.NewLogger(fmt.Sprintf("%s.%s", namespace, pipeline), &helper.standalone); err == nil {
			helper.logger = logger
		}
	}

	return helper
}

func (helper StandardHelper) IsDebugEnabled() bool {
	// TODO : add debug check
	if helper.standalone {
		return true
	}
	return api.IsDebugEnabled(helper.host)
}

func (helper StandardHelper) SaveToCache(data string) (string, error) {
	// TODO : add caching support for standalone mode
	if helper.standalone {
		helper.Warning(CachingInStandaloneErr.Error())
		return "", CachingInStandaloneErr
	}
	return api.Cache(helper.host, "", data)
}

func (helper StandardHelper) LoadFromCache(identifier string) (string, error) {
	// TODO : add caching support for standalone mode
	if helper.standalone {
		helper.Warning(CachingInStandaloneErr.Error())
		return "", CachingInStandaloneErr
	}
	return api.GetFromCache(helper.host, identifier)
}

func (helper StandardHelper) Logf(format string, data ...any) (err error) {
	output := fmt.Sprintf(format, data...)
	return helper.Log(output)
}

func (helper StandardHelper) Log(message string) (err error) {

	if helper.standalone {
		helper.logger.Println(message)
		err = nil
	} else {
		err = api.Log(helper.host, helper.supervisor, message)
		//fmt.Println(message)
	}

	return err
}

func (helper StandardHelper) Warningf(format string, data ...any) (err error) {

	output := fmt.Sprintf(format, data...)
	return helper.Warning(output)
}

func (helper StandardHelper) Warning(message string) (err error) {

	if helper.standalone {
		helper.logger.Warnln(message)
		err = nil
	} else {
		err = api.LogWarn(helper.host, helper.supervisor, message)
		//fmt.Println(message)
	}

	return err
}

func (helper StandardHelper) Fatalf(format string, data ...any) (err error) {

	output := fmt.Sprintf(format, data...)
	return helper.Fatal(output)
}

func (helper StandardHelper) Fatal(message string) (err error) {

	if helper.standalone {
		helper.logger.Alertln(message)
		err = nil
	} else {
		err = api.LogError(helper.host, helper.supervisor, message)
		//fmt.Println(message)
	}

	return err
}
