package llm

import "errors"

// ErrToolsUnsupported is returned by CallTools when the backend model does not
// support native tool calling, so callers can fall back to prompt-based use.
var ErrToolsUnsupported = errors.New("llm: model does not support tools")
