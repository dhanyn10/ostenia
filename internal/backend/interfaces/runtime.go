package interfaces

import (
	"context"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Runtime interface {
	EventsEmit(ctx context.Context, eventName string, optionalData ...interface{})
	WindowMinimise(ctx context.Context)
	WindowMaximise(ctx context.Context)
	WindowUnmaximise(ctx context.Context)
	WindowExecJS(ctx context.Context, js string)
	Quit(ctx context.Context)
	OpenFileDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error)
	OpenDirectoryDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error)
	SaveFileDialog(ctx context.Context, options wruntime.SaveDialogOptions) (string, error)
}
