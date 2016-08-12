package oracle

import (
    "net/http"
    "github.com/graymeta/stow"
)

type ConnectionInfo struct {
    client http.Client
    AuthInfo AuthResponse
}

type location struct {
    config stow.Config
    client ConnectionInfo
}
