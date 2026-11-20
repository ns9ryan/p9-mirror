package config

import (
	"github.com/zeromicro/go-zero/core/conf"
	"testing"
)

func TestMemberExportConfigurationOptional(t *testing.T) {
	type wrapper struct {
		MemberExport MemberExportConfig `json:",optional"`
	}
	for _, input := range []string{"{}", "MemberExport:\n  RedisAddr: localhost:6379\n"} {
		var value wrapper
		if err := conf.LoadFromYamlBytes([]byte(input), &value); err != nil {
			t.Fatalf("optional export config breaks existing startup: %v", err)
		}
	}
}
