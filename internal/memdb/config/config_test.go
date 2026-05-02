package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/niksmo/memdb/pkg/logger"
)

type mockYamlProvider struct {
	payload []byte
}

func (p *mockYamlProvider) Bytes() ([]byte, error) {
	return p.payload, nil
}

func TestLoad_Unmarshal_ValidConfig(t *testing.T) {
	t.Parallel()

	payload := []byte(`
engine:
  type: "test_type"
network:
  address: "127.0.0.1:3223"
  max_connections: 50
  max_message_size: 1024
  idle_timeout: 30s
logging:
  level: "warn"
  output: "/log/output.log"`)

	mock := &mockYamlProvider{payload}

	expected := Config{
		Engine: engine{Type: "test_type"},
		Network: network{
			Address:        "127.0.0.1:3223",
			MaxConnections: 50,
			MaxMessageSize: 1024,
			IdleTimeout:    30 * time.Second,
		},
		Logging: logging{
			Level:  "warn",
			Output: "/log/output.log",
		},
	}

	config := Load(os.Args[:1], WithFileProvider(mock))
	require.Equal(t, expected, config)
}

func TestLoad_Unmarshal_InvalidConfig(t *testing.T) {
	t.Parallel()

	payload := []byte(`
engine:
  kind: "in_our_memories"
network:
  address: "127.0.0.1:9999"
  protocol: "tcp"
max_connections: 50
max_message_size: 1024
idle_timeout: 30s
logging:
  level: "warn"
  out: "/log/output.log"`)

	mock := &mockYamlProvider{payload}

	expected := Config{
		Engine: engine{
			Type: "in_memory", // default
		},
		Network: network{
			Address:        "127.0.0.1:9999",
			MaxConnections: 100,             // default
			MaxMessageSize: 4 << 10,         // default
			IdleTimeout:    5 * time.Minute, // default,
		},
		Logging: logging{
			Level:  "warn",
			Output: "", // default
		},
	}

	config := Load(os.Args[:1], WithFileProvider(mock))
	require.Equal(t, expected, config)
}

func TestLoad_Unmarshal_BrokenConfig(t *testing.T) {
	t.Parallel()

	// broken part:
	//	`max_connections: 50
	//	  max_message_size: 1024`
	payload := []byte(`
engine:
  kind: "in_our_memories"
network:
  address: "127.0.0.1:9999"
  protocol: "tcp"
max_connections: 50
  max_message_size: 1024
  idle_timeout: 30s
      logging: "debug"
  level: "warn"
  out: "/log/output.log"`)

	mock := &mockYamlProvider{payload}

	expected := Config{
		Engine:  defaultEngine(),
		Network: defaultNetwork(),
		Logging: defaultLogging(),
	}

	observer := new(bytes.Buffer)
	l := logger.New(observer, "debug")
	config := Load(os.Args[:1], WithFileProvider(mock), WithLogger(l))
	require.Equal(t, expected, config)

	line, _ := observer.ReadString('\n')
	require.Contains(t, line, `msg="read configuration file" error`)
}

func TestFileProvider_DefaultPath(t *testing.T) {
	t.Parallel()

	payload := []byte(`
engine:
  type: "test_type"
network:
  address: "127.0.0.1:3223"
  max_connections: 50
  max_message_size: 1024
  idle_timeout: 30s
logging:
  level: "warn"
  output: "/log/output.log"`)

	rootPath := t.TempDir()
	f, err := os.Create(filepath.Join(rootPath, defaultPath))
	require.NoError(t, err)

	_, err = f.Write(payload)
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	provider := &yamlProvider{os.Args[:1], rootPath}
	yamlData, err := provider.Bytes()
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, payload, yamlData)
}

func TestFileProvider_AbsolutePath(t *testing.T) {
	t.Parallel()

	payload := []byte(`
engine:
  type: "test_type"
network:
  address: "127.0.0.1:3223"
  max_connections: 50
  max_message_size: 1024
  idle_timeout: 30s
logging:
  level: "warn"
  output: "/log/output.log"`)

	rootPath := t.TempDir()
	configPath := filepath.Join(rootPath, "abracadabra.yml")
	f, err := os.Create(configPath)
	require.NoError(t, err)

	_, err = f.Write(payload)
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	args := append(os.Args[:1], "-c", configPath)

	provider := &yamlProvider{args, rootPath}
	yamlData, err := provider.Bytes()
	require.NoError(t, err)
	require.Equal(t, payload, yamlData)
}
