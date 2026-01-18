package servers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	"github.com/pufferpanel/pufferpanel/v3"
	"github.com/pufferpanel/pufferpanel/v3/logging"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func CreateFunctions(env *pufferpanel.Environment) []cel.EnvOption {
	return []cel.EnvOption{
		cel.Function("file_exists",
			cel.Overload("file_exists_string_bool",
				[]*cel.Type{cel.StringType},
				cel.BoolType,
				cel.UnaryBinding(cel_file_exists(env)),
			)),
		cel.Function("in_path",
			cel.Overload("in_path_string_bool",
				[]*cel.Type{cel.StringType},
				cel.BoolType,
				cel.UnaryBinding(cel_in_path(env)),
			)),
		cel.Function("is_server_running",
			cel.Overload("is_server_running_bool",
				[]*cel.Type{},
				cel.BoolType,
				cel.FunctionBinding(cel_is_server_running(env)),
			)),
		cel.Function("file_hash",
			cel.Overload("file_hash_string_string",
				[]*cel.Type{cel.StringType},
				cel.StringType,
				cel.UnaryBinding(cel_file_hash(env)),
			)),
	}
}

func cel_file_exists(env *pufferpanel.Environment) functions.UnaryOp {
	return func(fileName ref.Val) ref.Val {
		fullPath := filepath.Join(env.GetRootDirectory(), fileName.Value().(string))
		_, err := os.Stat(fullPath)
		return types.Bool(err == nil)
	}
}

func cel_in_path(env *pufferpanel.Environment) functions.UnaryOp {
	return func(fileName ref.Val) ref.Val {
		_, err := exec.LookPath(fileName.Value().(string))
		return types.Bool(err == nil || errors.Is(err, exec.ErrDot))
	}
}

func cel_is_server_running(env *pufferpanel.Environment) functions.FunctionOp {
	return func(values ...ref.Val) ref.Val {
		r, err := env.IsRunning()
		return types.Bool(err == nil && r)
	}
}

func cel_file_hash(env *pufferpanel.Environment) functions.UnaryOp {
	return func(fileName ref.Val) ref.Val {
		fullPath := filepath.Join(env.GetRootDirectory(), fileName.Value().(string))
		_, err := os.Stat(fullPath)

		if err != nil {
			if os.IsNotExist(err) {
				logging.Error.Printf("Target file does not exist")
				return nil
			} else if !os.IsNotExist(err) {
				logging.Error.Printf("Error reading target file: %s", err)
				return nil
			}
		}

		file, err := os.Open(fullPath)
		if err != nil {
			logging.Error.Printf("Error opening target file: %s", err)
			return nil
		}
		defer file.Close()

		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return nil
		}
		fileHash := hex.EncodeToString(hasher.Sum(nil))
		return types.String(fileHash)
	}
}
