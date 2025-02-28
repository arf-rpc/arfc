package common

import (
	"github.com/arf-rpc/arfc/arf/strcase"
	"github.com/arf-rpc/idl/ast"
	"slices"
	"strings"
)

func CanonicalStructName(pkg string, s *ast.Struct) string {
	names := []string{s.Name}
	p := s.Parent
	for p != nil {
		names = append(names, p.Name)
		p = p.Parent
	}
	slices.Reverse(names)
	for i, v := range names {
		names[i] = strcase.ToSnake(v)
	}
	return pkg + "/" + strings.Join(names, "/")
}
