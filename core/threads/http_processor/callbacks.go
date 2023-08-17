package http_processor

import (
	"encoding/json"
	"errors"
	"github.com/GabeCordo/etl-light/module"
	"github.com/GabeCordo/etl-light/processor_i"
	"github.com/GabeCordo/etl/core/components/processor"
	"github.com/GabeCordo/etl/core/threads/common"
	"github.com/GabeCordo/etl/core/utils"
	"net/http"
	"net/url"
)

func (thread *Thread) processorCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	if r.Method == "POST" {
		cfg := &processor_i.Config{}
		if err := json.NewDecoder(r.Body).Decode(cfg); err == nil {
			success, err := common.AddProcessor(thread.C12, thread.ProcessorResponseTable, cfg)
			if !success && errors.Is(err, processor.AlreadyExists) {
				w.WriteHeader(http.StatusConflict)
			} else if !success {
				w.WriteHeader(http.StatusInternalServerError)
			}

			if err != nil {
				w.Write([]byte(err.Error()))
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else if r.Method == "DELETE" {

		processorName, processorNameFound := urlMapping["processor"]
		if !processorNameFound {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := common.DeleteProcessor(thread.C12, thread.ProcessorResponseTable, processorName[0])
		if errors.Is(err, processor.DoesNotExist) {
			w.WriteHeader(http.StatusNotFound)
		} else if errors.Is(err, utils.NoResponseReceived) {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if err != nil {
			w.Write([]byte(err.Error()))
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) moduleCallback(w http.ResponseWriter, r *http.Request) {

	urlMapping, _ := url.ParseQuery(r.URL.RawQuery)

	processorName, processorNameFound := urlMapping["processor"]

	if !processorNameFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		cfg := &module.Config{}
		if err := json.NewDecoder(r.Body).Decode(cfg); err == nil {
			success, err := common.AddModule(thread.C12, thread.ProcessorResponseTable, processorName[0], cfg)
			if !success && errors.Is(err, processor.ModuleAlreadyRegistered) {
				/* the module is already registered to the processor */
				w.WriteHeader(http.StatusConflict)
			} else if !success && errors.Is(err, processor.DoesNotExist) {
				/* the processor does not exist and can not bind a module */
				w.WriteHeader(http.StatusBadRequest)
			} else if !success {
				w.WriteHeader(http.StatusBadRequest)
			}

			if err != nil {
				w.Write([]byte(err.Error()))
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	} else if r.Method == "DELETE" {

		processorName, processorNameFound := urlMapping["processor"]
		moduleName, moduleNameFound := urlMapping["module"]

		if !processorNameFound || !moduleNameFound {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err := common.DeleteModule(thread.C12, thread.ProcessorResponseTable, processorName[0], moduleName[0])
		if errors.Is(err, processor.DoesNotExist) || errors.Is(err, processor.ModuleDoesNotExist) {
			w.WriteHeader(http.StatusNotFound)
		} else if errors.Is(err, utils.NoResponseReceived) {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if err != nil {
			w.Write([]byte(err.Error()))
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (thread *Thread) debugCallback(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		// treat this as a probe to the server
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
