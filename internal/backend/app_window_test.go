package backend

import (
	"context"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"ostenia/internal/backend/interfaces"
	"testing"
)

type FullMockRuntime struct {
	interfaces.Runtime
	Minimised   bool
	Maximised   bool
	Unmaximised bool
	QuitCalled  bool
	JS          string
	LastCtx     context.Context
}

func (m *FullMockRuntime) WindowMinimise(ctx context.Context) {
	m.Minimised = true
	m.LastCtx = ctx
}
func (m *FullMockRuntime) WindowMaximise(ctx context.Context) {
	m.Maximised = true
	m.LastCtx = ctx
}
func (m *FullMockRuntime) WindowUnmaximise(ctx context.Context) {
	m.Unmaximised = true
	m.LastCtx = ctx
}
func (m *FullMockRuntime) WindowExecJS(ctx context.Context, js string) {
	m.JS = js
	m.LastCtx = ctx
}
func (m *FullMockRuntime) Quit(ctx context.Context) {
	m.QuitCalled = true
	m.LastCtx = ctx
}
func (m *FullMockRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {
}
func (m *FullMockRuntime) OpenFileDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return "file", nil
}
func (m *FullMockRuntime) OpenDirectoryDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return "dir", nil
}
func (m *FullMockRuntime) SaveFileDialog(ctx context.Context, options wruntime.SaveDialogOptions) (string, error) {
	return "save", nil
}

func TestApp_WindowDelegates(t *testing.T) {
	mockR := &FullMockRuntime{}
	ctx := context.Background()
	app := &App{
		runtime: mockR,
	}
	app.Startup(ctx)

	app.Minimize()
	if !mockR.Minimised {
		t.Error("Expected WindowMinimise to be called")
	}

	app.Maximize()
	if !mockR.Maximised {
		t.Error("Expected WindowMaximise to be called")
	}

	app.Unmaximize()
	if !mockR.Unmaximised {
		t.Error("Expected WindowUnmaximise to be called")
	}

	app.Close()
	if !mockR.QuitCalled {
		t.Error("Expected Quit to be called")
	}

	app.ToggleDevTools()
	if mockR.JS == "" {
		t.Error("Expected WindowExecJS to be called")
	}

	app.EventsEmit(context.Background(), "test")
	app.Quit(context.Background())
	_, _ = app.OpenFileDialog(context.Background(), wruntime.OpenDialogOptions{})
	_, _ = app.OpenDirectoryDialog(context.Background(), wruntime.OpenDialogOptions{})
	_, _ = app.SaveFileDialog(context.Background(), wruntime.SaveDialogOptions{})
}

func TestApp_WindowDelegatesWithStartupContext(t *testing.T) {
	mockR := &FullMockRuntime{}
	type contextKey string
	startupCtx := context.WithValue(context.Background(), contextKey("type"), "startup")

	app := &App{
		runtime: mockR,
	}
	app.Startup(startupCtx)

	// Minimize
	app.Minimize()
	if !mockR.Minimised {
		t.Error("Expected WindowMinimise to be called")
	}
	if mockR.LastCtx != startupCtx {
		t.Errorf("Expected WindowMinimise to use startupCtx, got: %v", mockR.LastCtx)
	}

	// Maximize
	app.Maximize()
	if !mockR.Maximised {
		t.Error("Expected WindowMaximise to be called")
	}
	if mockR.LastCtx != startupCtx {
		t.Errorf("Expected WindowMaximise to use startupCtx, got: %v", mockR.LastCtx)
	}

	// Unmaximize
	app.Unmaximize()
	if !mockR.Unmaximised {
		t.Error("Expected WindowUnmaximise to be called")
	}
	if mockR.LastCtx != startupCtx {
		t.Errorf("Expected WindowUnmaximise to use startupCtx, got: %v", mockR.LastCtx)
	}

	// Close
	app.Close()
	if !mockR.QuitCalled {
		t.Error("Expected Quit to be called")
	}
	if mockR.LastCtx != startupCtx {
		t.Errorf("Expected Quit to use startupCtx, got: %v", mockR.LastCtx)
	}

	// ToggleDevTools
	app.ToggleDevTools()
	if mockR.JS == "" {
		t.Error("Expected WindowExecJS to be called")
	}
	if mockR.LastCtx != startupCtx {
		t.Errorf("Expected WindowExecJS to use startupCtx, got: %v", mockR.LastCtx)
	}
}
