package screenshot

import (
	"fmt"
	"net/url"
	"path"
)

type URLFormatter interface {
	Format(push, pull *url.URL, id string, useStrftime bool) string
}

type SimpleURLFormatter struct{}

func (sf *SimpleURLFormatter) Format(push, _ *url.URL, id string, _ bool) string {
	return path.Join(push.Path, fmt.Sprintf("%s.jpg", id))
}
