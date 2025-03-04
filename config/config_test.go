package config

import (
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func Test_Defaults(t *testing.T) {
	tcs := []struct {
		path          string
		expectedValue interface{}
	}{
		{
			path:          "L1.RpcURL",
			expectedValue: "ws://127.0.0.1:8546",
		},
		{
			path:          "L1.PolygonValidiumAddress",
			expectedValue: "0x8dAF17A20c9DBA35f005b6324F493785D239719d",
		},
		{
			path:          "L1.Timeout",
			expectedValue: types.NewDuration(1 * time.Minute),
		},
		{
			path:          "L1.RetryPeriod",
			expectedValue: types.NewDuration(5 * time.Second),
		},
		{
			path:          "L1.BlockBatchSize",
			expectedValue: uint(64),
		},
		{
			path:          "PermitApiAddress",
			expectedValue: common.Address{},
		},
		// TODO: more default checks
	}

	ctx := cli.NewContext(cli.NewApp(), nil, nil)
	cfg, err := Load(ctx)
	require.NoError(t, err)

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			actual := getValueFromStruct(tc.path, cfg)
			require.Equal(t, tc.expectedValue, actual)
		})
	}
}

func Test_ConfigFileNotFound(t *testing.T) {
	flags := flag.FlagSet{}
	flags.String("cfg", "/fictitious-file/foo.cfg", "")

	ctx := cli.NewContext(cli.NewApp(), &flags, nil)
	_, err := Load(ctx)
	require.Error(t, err)
}

func Test_ConfigFileOverride(t *testing.T) {
	tempDir := t.TempDir()
	overrides := filepath.Join(tempDir, "overrides.toml")
	f, err := os.Create(overrides)
	require.NoError(t, err)
	_, err = f.WriteString("[L1]\n")
	require.NoError(t, err)
	_, err = f.WriteString("PolygonValidiumAddress = \"0xDEADBEEF\"")
	require.NoError(t, err)
	flags := flag.FlagSet{}
	flags.String("cfg", overrides, "")
	ctx := cli.NewContext(cli.NewApp(), &flags, nil)
	cfg, err := Load(ctx)
	require.NoError(t, err)
	require.Equal(t, "0xDEADBEEF", cfg.L1.PolygonValidiumAddress)
}

func Test_NewKeyFromKeystore(t *testing.T) {
	t.Parallel()

	t.Run("valid keystore file", func(t *testing.T) {
		t.Parallel()

		cfg := types.KeystoreFileConfig{
			Path:     "../test/config/test-member.keystore",
			Password: "testonly",
		}

		key, err := NewKeyFromKeystore(cfg)
		require.NoError(t, err)
		require.NotNil(t, key)
	})

	t.Run("no path and password", func(t *testing.T) {
		t.Parallel()

		cfg := types.KeystoreFileConfig{}

		key, err := NewKeyFromKeystore(cfg)
		require.NoError(t, err)
		require.Nil(t, key)
	})

	t.Run("invalid keystore file", func(t *testing.T) {
		t.Parallel()

		cfg := types.KeystoreFileConfig{
			Path:     "non-existent.keystore",
			Password: "testonly",
		}

		key, err := NewKeyFromKeystore(cfg)
		require.ErrorContains(t, err, "no such file or directory")
		require.Nil(t, key)
	})

	t.Run("invalid password", func(t *testing.T) {
		t.Parallel()

		cfg := types.KeystoreFileConfig{
			Path:     "../test/config/test-member.keystore",
			Password: "invalid",
		}

		key, err := NewKeyFromKeystore(cfg)
		require.ErrorContains(t, err, "could not decrypt key with given password")
		require.Nil(t, key)
	})
}

func getValueFromStruct(path string, object interface{}) interface{} {
	keySlice := strings.Split(path, ".")
	v := reflect.ValueOf(object)

	for _, key := range keySlice {
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		v = v.FieldByName(key)
	}
	return v.Interface()
}
