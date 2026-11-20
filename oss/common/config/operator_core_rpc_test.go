package config

import "testing"

func TestOperatorCoreRPCDoesNotReplaceCoreRPC(t *testing.T) {
	config := unmarshalYAML(t, `
port: "8020"
core_rpc:
  endpoints: [platform-core-rpc:19000]
  timeout: 10000
operator_core_rpc:
  endpoints: [operator-core-rpc:19000]
  timeout: 8000
oss:
  buckets:
    images:
      mode: public
`)
	if len(config.CoreRPC.Endpoints) != 1 || config.CoreRPC.Endpoints[0] != "platform-core-rpc:19000" {
		t.Fatal("original Core config changed")
	}
	if len(config.OperatorCoreRPC.Endpoints) != 1 || config.OperatorCoreRPC.Endpoints[0] != "operator-core-rpc:19000" || config.OperatorCoreRPC.Timeout != 8000 {
		t.Fatal("operator Core config missing")
	}
}
