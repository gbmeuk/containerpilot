package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/joyent/containerpilot/tests"
	"github.com/joyent/containerpilot/tests/assert"
	"github.com/joyent/containerpilot/tests/mocks"
)

var noop = &mocks.NoopDiscoveryBackend{}

func TestServiceConfigHappyPath(t *testing.T) {

	testCfg := tests.DecodeRawToSlice(`[
	{
		"name": "serviceA",
		"port": 8080,
		"interfaces": "inet",
		"exec": "/bin/serviceA",
		"preStart": "/bin/to/preStart.sh arg1 arg2",
		"preStop": ["/bin/to/preStop.sh","arg1","arg2"],
		"postStop": ["/bin/to/postStop.sh"],
		"health": "/bin/to/healthcheck/for/service/A.sh",
		"poll": 30,
		"ttl": "19",
		"tags": ["tag1","tag2"]
	},
	{
		"name": "serviceB",
		"port": 5000,
		"interfaces": ["ethwe","eth0", "inet"],
		"exec": ["/bin/serviceB", "B"],
		"health": ["/bin/to/healthcheck/for/service/B.sh", "B"],
		"timeout": "2s",
		"poll": 20,
		"ttl": "103"
	},
	{
		"name": "coprocessC",
		"exec": "/bin/coprocessC",
		"restarts": "unlimited"
	},
		{
		"name": "taskD",
		"exec": "/bin/taskD",
		"frequency": "1s"
	}
]
`)

	services, err := NewConfigs(testCfg, noop)
	if err != nil {
		t.Fatalf("unexpected error in LoadConfig: %v", err)
	}

	if len(services) != 7 {
		t.Fatalf("expected 7 services but got %v", services)
	}
	svc0 := services[0]
	assert.Equal(t, svc0.Name, "serviceA", "expected '%v' for svc0.Name but got '%v'")
	assert.Equal(t, svc0.Port, 8080, "expected '%v' for svc0.Port but got '%v'")
	assert.Equal(t, svc0.Exec, "/bin/serviceA", "expected '%v' for svc0.Exec but got '%v'")
	assert.Equal(t, svc0.Tags, []string{"tag1", "tag2"}, "expected '%v' for svc0.Tags but got '%v'")
	assert.Equal(t, svc0.Restarts, nil, "expected '%v' for svc1.Restarts but got '%v'")

	svc1 := services[1]
	assert.Equal(t, svc1.Name, "serviceB", "expected '%v' for svc1.Name but got '%v'")
	assert.Equal(t, svc1.Port, 5000, "expected '%v' for svc1.Port but got '%v'")
	assert.Equal(t, len(svc1.Tags), 0, "expected '%v' for len(svc1.Tags) but got '%v'")
	assert.Equal(t, svc1.Exec, []interface{}{"/bin/serviceB", "B"}, "expected '%v' for svc1.Exec but got '%v'")
	assert.Equal(t, svc1.Restarts, nil, "expected '%v' for svc1.Restarts but got '%v'")

	svc2 := services[2]
	assert.Equal(t, svc2.Name, "coprocessC", "expected '%v' for svc2.Name but got '%v'")
	assert.Equal(t, svc2.Port, 0, "expected '%v' for svc2.Port but got '%v'")
	assert.Equal(t, svc2.Frequency, "", "expected '%v' for svc2.Frequency but got '%v'")
	assert.Equal(t, svc2.Restarts, "unlimited", "expected '%v' for svc2.Restarts but got '%v'")

	svc3 := services[3]
	assert.Equal(t, svc3.Name, "taskD", "expected '%v' for svc3.Name but got '%v'")
	assert.Equal(t, svc3.Port, 0, "expected '%v' for svc3.Port but got '%v'")
	assert.Equal(t, svc3.Frequency, "1s", "expected '%v' for svc3.Frequency but got '%v'")
	assert.Equal(t, svc3.Restarts, nil, "expected '%v' for svc3.Restarts but got '%v'")

	svc4 := services[4]
	assert.Equal(t, svc4.Name, "serviceA.preStart", "expected '%v' for svc4.Name but got '%v'")
	assert.Equal(t, svc4.Port, 0, "expected '%v' for svc4.Port but got '%v'")
	assert.Equal(t, svc4.Frequency, "", "expected '%v' for svc4.Frequency but got '%v'")
	assert.Equal(t, svc4.Restarts, nil, "expected '%v' for svc4.Restarts but got '%v'")

	svc5 := services[5]
	assert.Equal(t, svc5.Name, "serviceA.preStop", "expected '%v' for svc5.Name but got '%v'")
	assert.Equal(t, svc5.Port, 0, "expected '%v' for svc5.Port but got '%v'")
	assert.Equal(t, svc5.Frequency, "", "expected '%v' for svc5.Frequency but got '%v'")
	assert.Equal(t, svc5.Restarts, nil, "expected '%v' for svc5.Restarts but got '%v'")

	svc6 := services[6]
	assert.Equal(t, svc6.Name, "serviceA.postStop", "expected '%v' for svc6.Name but got '%v'")
	assert.Equal(t, svc6.Port, 0, "expected '%v' for svc6.Port but got '%v'")
	assert.Equal(t, svc6.Frequency, "", "expected '%v' for svc6.Frequency but got '%v'")
	assert.Equal(t, svc6.Restarts, nil, "expected '%v' for svc6.Restarts but got '%v'")
}

func TestServiceConfigValidateName(t *testing.T) {

	_, err := NewConfigs(tests.DecodeRawToSlice(`[{"name": ""}]`), noop)
	assert.Error(t, err, "`name` must not be blank")

	cfg, err := NewConfigs(tests.DecodeRawToSlice(`[{"name": "", "exec": "myexec"}]`), noop)
	assert.Error(t, err, "`name` must not be blank")

	// no name permitted only if no discovery backend assigned
	cfg, err = NewConfigs(tests.DecodeRawToSlice(`[{"name": "", "exec": "myexec"}]`), nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cfg[0].Name, "myexec", "expected '%v' for cfg.Name got '%v'")
}

func TestServiceConfigValidateDiscovery(t *testing.T) {
	_, err := NewConfigs(tests.DecodeRawToSlice(`[{"name": "myName", "port": 80}]`), noop)
	assert.Error(t, err, "`poll` must be > 0 in service `myName` when `port` is set")

	_, err = NewConfigs(tests.DecodeRawToSlice(`[{"name": "myName", "port": 80, "poll": 1}]`), noop)
	assert.Error(t, err, "`ttl` must be > 0 in service `myName` when `port` is set")

	_, err = NewConfigs(tests.DecodeRawToSlice(`[{"name": "myName", "poll": 1, "ttl": 1}]`), noop)
	assert.Error(t, err,
		"`heartbeat` and `ttl` may not be set in service `myName` if `port` is not set")

	// no health check shouldn't return an error
	raw := tests.DecodeRawToSlice(`[{"name": "myName", "poll": 1, "ttl": 1, "port": 80}]`)
	if _, err = NewConfigs(raw, noop); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
}

func TestServicesConsulExtrasEnableTagOverride(t *testing.T) {
	testCfg := `[
	{
	  "name": "serviceA",
	  "port": 8080,
	  "interfaces": "inet",
	  "health": ["/bin/to/healthcheck/for/service/A.sh", "A1", "A2"],
	  "poll": 30,
	  "ttl": 19,
	  "timeout": "1ms",
	  "tags": ["tag1","tag2"],
	  "consul": {
		  "enableTagOverride": true
	  }
	}
	]`

	if services, err := NewConfigs(tests.DecodeRawToSlice(testCfg), nil); err != nil {
		t.Fatalf("could not parse service JSON: %s", err)
	} else {
		if services[0].definition.ConsulExtras.EnableTagOverride != true {
			t.Errorf("ConsulExtras should have had EnableTagOverride set to true.")
		}
	}
}

func TestInvalidServicesConsulExtrasEnableTagOverride(t *testing.T) {
	testCfg := `[
	{
	  "name": "serviceA",
	  "port": 8080,
	  "interfaces": "inet",
	  "health": ["/bin/to/healthcheck/for/service/A.sh", "A1", "A2"],
	  "poll": 30,
	  "ttl": 19,
	  "timeout": "1ms",
	  "tags": ["tag1","tag2"],
	  "consul": {
		  "enableTagOverride": "nope"
	  }
	}
	]`

	if _, err := NewConfigs(tests.DecodeRawToSlice(testCfg), nil); err == nil {
		t.Errorf("ConsulExtras should have thrown error about EnableTagOverride being a string.")
	}
}

func TestServicesConsulExtrasDeregisterCriticalServiceAfter(t *testing.T) {
	testCfg := `[
	{
	  "name": "serviceA",
	  "port": 8080,
	  "interfaces": "inet",
	  "health": ["/bin/to/healthcheck/for/service/A.sh", "A1", "A2"],
	  "poll": 30,
	  "ttl": 19,
	  "timeout": "1ms",
	  "tags": ["tag1","tag2"],
	  "consul": {
		  "deregisterCriticalServiceAfter": "40m"
	  }
	}
	]`

	if services, err := NewConfigs(tests.DecodeRawToSlice(testCfg), nil); err != nil {
		t.Fatalf("could not parse service JSON: %s", err)
	} else {
		if services[0].definition.ConsulExtras.DeregisterCriticalServiceAfter != "40m" {
			t.Errorf("ConsulExtras should have had DeregisterCriticalServiceAfter set to '40m'.")
		}
	}
}

func TestInvalidServicesConsulExtrasDeregisterCriticalServiceAfter(t *testing.T) {
	testCfg := `[
	{
	  "name": "serviceA",
	  "port": 8080,
	  "interfaces": "inet",
	  "health": ["/bin/to/healthcheck/for/service/A.sh", "A1", "A2"],
	  "poll": 30,
	  "ttl": 19,
	  "timeout": "1ms",
	  "tags": ["tag1","tag2"],
	  "consul": {
		  "deregisterCriticalServiceAfter": "nope"
	  }
	}
	]`

	if _, err := NewConfigs(tests.DecodeRawToSlice(testCfg), nil); err == nil {
		t.Errorf("error should have been generated for duration 'nope'.")
	}
}

func TestServiceConfigValidateFrequency(t *testing.T) {
	expectErr := func(test, errMsg string) {
		testCfg := tests.DecodeRawToSlice(test)
		_, err := NewConfigs(testCfg, nil)
		assert.Error(t, err, errMsg)
	}
	expectErr(`[{"exec": "/bin/taskA", "frequency": "-1s", "execTimeout": "1s"}]`,
		"frequency '-1s' cannot be less than 1ms")

	expectErr(`[{"exec": "/bin/taskB", "frequency": "1ns", "execTimeout": "1s"}]`,
		"frequency '1ns' cannot be less than 1ms")

	expectErr(`[{"exec": "/bin/taskC", "frequency": "1ms", "execTimeout": "-1ms"}]`,
		"timeout '-1ms' cannot be less than 1ms")

	expectErr(`[{"exec": "/bin/taskD", "frequency": "1ms", "execTimeout": "1ns"}]`,
		"timeout '1ns' cannot be less than 1ms")

	expectErr(`[{"exec": "/bin/taskD", "frequency": "xx", "execTimeout": "1ns"}]`,
		"unable to parse frequency 'xx': time: invalid duration xx")

	testCfg := tests.DecodeRawToSlice(`[{"exec": "/bin/taskE", "frequency": "1ms"}]`)
	service, _ := NewConfigs(testCfg, nil)
	assert.Equal(t, service[0].execTimeout, service[0].freqInterval,
		"expected execTimeout '%v' to equal frequency '%v'")
}

func TestServiceConfigValidateExec(t *testing.T) {

	testCfg := tests.DecodeRawToSlice(`[
	{
		"name": "serviceA",
		"exec": ["/bin/serviceA", "A1", "A2"],
		"health": ["/bin/to/healthcheck/for/service/A.sh", "A1", "A2"],
		"execTimeout": "1ms"
	}]`)
	cfg, err := NewConfigs(testCfg, noop)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cfg[0].exec.Exec, "/bin/serviceA",
		"expected %v for serviceA.exec.Exec got %v")
	assert.Equal(t, cfg[0].exec.Args, []string{"A1", "A2"},
		"expected %v for serviceA.exec.Args got %v")
	assert.Equal(t, cfg[0].execTimeout, time.Duration(time.Millisecond),
		"expected %v for serviceA.execTimeout got %v")

	testCfg = tests.DecodeRawToSlice(`[
	{
		"name": "serviceB",
		"exec": "/bin/serviceB B1 B2",
		"health": "/bin/to/healthcheck/for/service/B.sh B1 B2",
		"execTimeout": "1ms"
	}]`)
	cfg, err = NewConfigs(testCfg, noop)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cfg[0].exec.Exec, "/bin/serviceB",
		"expected %v for serviceB.exec.Exec got %v")
	assert.Equal(t, cfg[0].exec.Args, []string{"B1", "B2"},
		"expected %v for serviceB.exec.Args got %v")
	assert.Equal(t, cfg[0].execTimeout, time.Duration(time.Millisecond),
		"expected %v for serviceB.execTimeout got %v")

	testCfg = tests.DecodeRawToSlice(`[
	{
		"name": "serviceC",
		"exec": "/bin/serviceC C1 C2",
		"execTimeout": "xx"
	}]`)
	_, err = NewConfigs(testCfg, noop)
	expected := "could not parse `timeout` for service serviceC: time: invalid duration xx"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected '%s', got '%v'", expected, err)
	}

	testCfg = tests.DecodeRawToSlice(`[
	{
		"name": "serviceD",
		"exec": ""
	}]`)
	_, err = NewConfigs(testCfg, noop)
	expected = "could not parse `exec` for service serviceD: received zero-length argument"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected '%s', got '%v'", expected, err)
	}

}

func TestServiceConfigValidateRestarts(t *testing.T) {

	expectErr := func(test, val string) {
		errMsg := fmt.Sprintf(`invalid 'restarts' field "%v": accepts positive integers, "unlimited", or "never"`, val)
		testCfg := tests.DecodeRawToSlice(test)
		_, err := NewConfigs(testCfg, nil)
		assert.Error(t, err, errMsg)
	}
	expectErr(`[{"exec": "/bin/coprocessA", "restarts": "invalid"}]`, "invalid")
	expectErr(`[{"exec": "/bin/coprocessB", "restarts": "-1"}]`, "-1")
	expectErr(`[{"exec": "/bin/coprocessC", "restarts": -1 }]`, "-1")

	testCfg := tests.DecodeRawToSlice(`[
	{ "exec": "/bin/coprocessD", "restarts": "unlimited" },
	{ "exec": "/bin/coprocessE", "restarts": "never" },
	{ "exec": "/bin/coprocessF", "restarts": 1 },
	{ "exec": "/bin/coprocessG", "restarts": "1" },
	{ "exec": "/bin/coprocessH", "restarts": 0 },
	{ "exec": "/bin/coprocessI", "restarts": "0" },
	{ "exec": "/bin/coprocessJ"}
]
`)
	cfg, _ := NewConfigs(testCfg, nil)
	expectMsg := "expected restarts=%v got %v"

	assert.Equal(t, cfg[0].restartLimit, -1, expectMsg)
	assert.Equal(t, cfg[1].restartLimit, 0, expectMsg)
	assert.Equal(t, cfg[2].restartLimit, 1, expectMsg)
	assert.Equal(t, cfg[3].restartLimit, 1, expectMsg)
	assert.Equal(t, cfg[4].restartLimit, 0, expectMsg)
	assert.Equal(t, cfg[5].restartLimit, 0, expectMsg)
	assert.Equal(t, cfg[6].restartLimit, 0, expectMsg)
}

func TestServiceConfigPreStart(t *testing.T) {
	testCfg := tests.DecodeRawToSlice(`[
	{
		"name": "serviceA",
		"exec": "/bin/serviceA",
		"preStart": "/bin/to/preStart.sh arg1 arg2"
	}]`)
	cfg, err := NewConfigs(testCfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cfg[1].Name, "serviceA.preStart", "expected '%v' for preStart.Name got '%v")
	assert.Equal(t, cfg[1].exec.Exec, "/bin/to/preStart.sh",
		"expected '%v' for preStart.exec.Exec got '%v")

	testCfg = tests.DecodeRawToSlice(`[
	{
		"name": "serviceB",
		"exec": "/bin/serviceB",
		"preStart": ""
	}]`)
	_, err = NewConfigs(testCfg, nil)
	assert.Error(t, err,
		"could not parse `exec` for service serviceB.preStart: received zero-length argument")
}

func TestServiceConfigPreStop(t *testing.T) {
	testCfg := tests.DecodeRawToSlice(`[
	{
		"name": "serviceA",
		"exec": "/bin/serviceA",
		"preStop": "/bin/to/preStop.sh arg1 arg2"
	}]`)
	cfg, err := NewConfigs(testCfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cfg[1].Name, "serviceA.preStop", "expected '%v' for preStop.Name got '%v")
	assert.Equal(t, cfg[1].exec.Exec, "/bin/to/preStop.sh",
		"expected '%v' for preStop.exec.Exec got '%v")

	testCfg = tests.DecodeRawToSlice(`[
	{
		"name": "serviceB",
		"exec": "/bin/serviceB",
		"preStop": ""
	}]`)
	_, err = NewConfigs(testCfg, nil)
	assert.Error(t, err,
		"could not parse `exec` for service serviceB.preStop: received zero-length argument")
}

func TestServiceConfigPostStop(t *testing.T) {
	testCfg := tests.DecodeRawToSlice(`[
	{
		"name": "serviceA",
		"exec": "/bin/serviceA",
		"postStop": "/bin/to/postStop.sh arg1 arg2"
	}]`)
	cfg, err := NewConfigs(testCfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cfg[1].Name, "serviceA.postStop", "expected '%v' for postStop.Name got '%v")
	assert.Equal(t, cfg[1].exec.Exec, "/bin/to/postStop.sh",
		"expected '%v' for postStop.exec.Exec got '%v")

	testCfg = tests.DecodeRawToSlice(`[
	{
		"name": "serviceB",
		"exec": "/bin/serviceB",
		"postStop": ""
	}]`)
	_, err = NewConfigs(testCfg, nil)
	assert.Error(t, err,
		"could not parse `exec` for service serviceB.postStop: received zero-length argument")
}
