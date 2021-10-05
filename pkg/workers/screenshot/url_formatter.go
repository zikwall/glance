package screenshot

import (
	"fmt"
	"net/url"
	"path"
)

type UrlFormatter interface {
	Format(push, pull *url.URL, id string, useStrftime bool) string
}

type SimpleUrlFormatter struct{}

func (sf *SimpleUrlFormatter) Format(push, _ *url.URL, id string, _ bool) string {
	return path.Join(push.Path, fmt.Sprintf("%s.jpg", id))
}
