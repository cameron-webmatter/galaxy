package lifecycle

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
)

type FunctionHook struct {
	onStartup  func() error
	onShutdown func() error
}

func (f *FunctionHook) OnStartup() error {
	if f.onStartup != nil {
		return f.onStartup()
	}
	return nil
}

func (f *FunctionHook) OnShutdown() error {
	if f.onShutdown != nil {
		return f.onShutdown()
	}
	return nil
}

func LoadFromDir(srcDir string) (LifecycleHook, error) {
	lifecyclePath := filepath.Join(srcDir, "lifecycle.go")

	if _, err := os.Stat(lifecyclePath); os.IsNotExist(err) {
		return nil, nil
	}

	return nil, fmt.Errorf("lifecycle loading requires build-time code generation")
}

func DetectLifecycle(srcDir string) bool {
	lifecyclePath := filepath.Join(srcDir, "lifecycle.go")
	_, err := os.Stat(lifecyclePath)
	return err == nil
}

func LoadFromPlugin(pluginPath string) (LifecycleHook, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("open plugin: %w", err)
	}

	lifecycleSym, err := p.Lookup("Lifecycle")
	if err != nil {
		startupSym, startupErr := p.Lookup("OnStartup")
		shutdownSym, shutdownErr := p.Lookup("OnShutdown")

		if startupErr != nil && shutdownErr != nil {
			return nil, fmt.Errorf("no lifecycle functions found in plugin")
		}

		hook := &FunctionHook{}

		if startupErr == nil {
			if fn, ok := startupSym.(func() error); ok {
				hook.onStartup = fn
			}
		}

		if shutdownErr == nil {
			if fn, ok := shutdownSym.(func() error); ok {
				hook.onShutdown = fn
			}
		}

		return hook, nil
	}

	if fn, ok := lifecycleSym.(func() LifecycleHook); ok {
		return fn(), nil
	}

	lifecycleValue := reflect.ValueOf(lifecycleSym).Elem()
	if lifecycleValue.Kind() == reflect.Func {
		results := lifecycleValue.Call(nil)
		if len(results) == 1 {
			if hook, ok := results[0].Interface().(LifecycleHook); ok {
				return hook, nil
			}
		}
	}

	return nil, fmt.Errorf("invalid lifecycle export")
}
