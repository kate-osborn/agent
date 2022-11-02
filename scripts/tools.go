//go:build tools
// +build tools
// https://marcofranssen.nl/manage-go-tools-via-go-modules

package tools

import (
	_ "github.com/alvaroloes/enumer"
	_ "github.com/gogo/protobuf/protoc-gen-gogo"
	_ "github.com/gogo/protobuf/protoc-gen-gogofast"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
	_ "github.com/mwitkow/go-proto-validators/protoc-gen-govalidators"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
	_ "github.com/goreleaser/nfpm/v2/cmd/nfpm"
)
