package main

import (
	"ETLFramework/frontend"
	"testing"
)

/**
 * \fn		TestMarshalJSONConfig
 * \brief	Function verifies that the ETL Package can properly Marshal config JSON
 *          data and store it within Node, Auth, and Logging Golang Structures
 * \note	If the config.frontend.json format has changed, the reflections must be made in
 * 			this test case.
 */
func TestMarshalJSONConfig(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Error("JSON Marshal failed to transform to a Golang Struct")
		}
	}()

	node := frontend.GetNodeInstance()
	if node.Name != "Template" {
		t.Error("JSON Name not reflected in Marshal")
	}
	if node.Debug != true {
		t.Error("JSON Debug not reflected in Marshal")
	}
	if node.Address.Host != "127.0.0.1" {
		t.Error("JSON Address Host not reflected in Marshal")
	}
	if node.Address.Port != 8000 {
		t.Error("JSON Address Port not reflected in Marshal")
	}
}
