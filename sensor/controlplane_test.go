package sensor

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kubescape/host-scanner/sensor/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_removeEncryptionProviderConfigSecrets(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name   string
		args   args
		output string
	}{
		{
			name: "test",
			args: args{
				data: `{
					"apiVersion": "apiserver.config.k8s.io/v1",
					"kind": "EncryptionConfiguration",
					"resources": [
					  {
						"providers": [
						  {
							"aescbc": {
							  "keys": [
								{
								  "name": "key1",
								  "secret": "<BASE 64 ENCODED SECRET>"
								}
							  ]
							}
						  },
						  {
							"identity": {}
						  }
						],
						"resources": [
						  "secrets"
						]
					  }
					]
				  }`,
			},
			output: `{
				"apiVersion": "apiserver.config.k8s.io/v1",
				"kind": "EncryptionConfiguration",
				"resources": [
				  {
					"providers": [
					  {
						"aescbc": {
						  "keys": [
							{
							  "name": "key1",
							  "secret": "<REDACTED>"
							}
						  ]
						}
					  },
					  {
						"identity": {}
					  }
					],
					"resources": [
					  "secrets"
					]
				  }
				]
			  }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data map[string]interface{}
			json.Unmarshal([]byte(tt.args.data), &data)
			removeEncryptionProviderConfigSecrets(data)
			v, _ := json.Marshal(data)
			require.JSONEq(t, tt.output, string(v))
		})
	}
}

func Test_makeAPIserverAuditPolicyFile(t *testing.T) {
	uid_tests := []struct {
		name     string
		cmdLine  string
		expected int
	}{
		{
			name:    "audit_policy_file_exist",
			cmdLine: "kube-apiserver --audit-policy-file=./testdata/audit-policy.yaml",
		},
		{
			name:    "audit_policy_file_exist",
			cmdLine: "kube-apiserver",
		},
	}

	for _, tt := range uid_tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			process := &utils.ProcessDetails{
				CmdLine: strings.Split(tt.cmdLine, " "),
			}
			fileInfo := makeAPIserverAuditPolicyFile(ctx, process)
			if assert.NotEqual(t, fileInfo, nil) {
				t.Log("err")
			}
		})
	}
}
