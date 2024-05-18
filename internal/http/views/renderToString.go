package views

import (
	"bytes"
	"context"

	"github.com/a-h/templ"
)

func RenderToString(t templ.Component) string {
	buf := bytes.NewBufferString("")
	t.Render(context.Background(), buf)

	return buf.String()
}
