package common

import (
	"fmt"
	"github.com/arf-rpc/arfc/arf/strcase"
	"github.com/arf-rpc/idl/ast"
	"slices"
	"strings"
)

type PackageResolver func(name string, typeName string) string

func MaybePointer(t ast.Type, resolver PackageResolver) string {
	str := ConvertType(t, resolver)
	if IsUserType(t) {
		str = "*" + str
	}
	return str
}

func PackageName(t ast.Object) string {
	comps := strings.Split(t.BaseFQN(), ".")
	return comps[len(comps)-1]
}

func ConvertType(t ast.Type, resolver PackageResolver) string {
	switch v := t.(type) {
	case *ast.PrimitiveType:
		switch v.Name {
		case "timestamp":
			return "time.Time"
		case "bytes":
			return "[]byte"
		default:
			return v.Name
		}
	case *ast.OptionalType:
		return "*" + ConvertType(v.Type, resolver)
	case *ast.ArrayType:
		return "[]" + ConvertType(v.Type, resolver)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", ConvertType(v.Key, resolver), ConvertType(v.Value, resolver))
	case *ast.SimpleUserType:
		switch a := v.ResolvedType.(type) {
		case *ast.Enum:
			return EnumName(a)
		case *ast.Struct:
			return StructName(a)
		default:
			return "INVALID"
		}
	case *ast.FullQualifiedType:
		prefix := ""
		if resolver != nil {
			p := resolver(v.Name, v.Package)
			if p != "" {
				prefix = p + "."
			}
		}
		switch a := v.ResolvedType.(type) {
		case *ast.Enum:
			return prefix + EnumName(a)
		case *ast.Struct:
			return prefix + StructName(a)
		default:
			return "INVALID"
		}
	default:
		return "INVALID"
	}
}

func EnumName(e *ast.Enum) string {
	names := []string{e.Name}
	p := e.Parent
	for p != nil {
		names = append(names, p.Name)
		p = p.Parent
	}
	slices.Reverse(names)
	for i, v := range names {
		names[i] = strcase.ToCamel(v)
	}
	return strings.Join(names, "")
}

func StructName(s *ast.Struct) string {
	names := []string{s.Name}
	p := s.Parent
	for p != nil {
		names = append(names, p.Name)
		p = p.Parent
	}
	slices.Reverse(names)
	for i, v := range names {
		names[i] = strcase.ToCamel(v)
	}
	return strings.Join(names, "")
}

func IsUserType(t ast.Type) bool {
	switch usr := t.(type) {
	case *ast.SimpleUserType:
		if _, ok := usr.ResolvedType.(*ast.Enum); !ok {
			return true
		}
	case *ast.FullQualifiedType:
		if _, ok := usr.ResolvedType.(*ast.Enum); !ok {
			return true
		}
	}
	return false
}

type MethodInput struct {
	Name string
	Type ast.Type
}
type MethodDefinition struct {
	Method           *ast.ServiceMethod
	Name             string
	Inputs           []MethodInput
	Output           []ast.Type
	HasInput         bool
	HasOutput        bool
	HasInputStream   bool
	HasOutputStream  bool
	InputStreamType  ast.Type
	OutputStreamType ast.Type
	Tree             *ast.PackageTree
}

func InStreamer(t ast.Type, resolver PackageResolver) string {
	return fmt.Sprintf("arf.InStreamer[%s]", MaybePointer(t, resolver))
}
func OutStreamer(t ast.Type, resolver PackageResolver) string {
	return fmt.Sprintf("arf.OutStreamer[%s]", MaybePointer(t, resolver))
}
func InOutStreamer(i, o ast.Type, resolver PackageResolver) string {
	return fmt.Sprintf("arf.InOutStreamer[%s, %s]", MaybePointer(i, resolver), MaybePointer(o, resolver))
}

func (m *MethodDefinition) ResponderName() string {
	return fmt.Sprintf("%s%sResponder", strcase.ToCamel(PackageName(m.Method)), strcase.ToCamel(m.Name))
}

func (m *MethodDefinition) BuildInterfaceSignature(w *Writer, resolver PackageResolver) {
	isNamed := m.HasInput

	w.Writef("%s(", strcase.ToCamel(m.Name))
	if isNamed {
		w.Writef("ctx ")
	}
	w.Writef("context.Context, ")
	switch {
	case !m.HasInput && !m.HasOutput && !m.HasInputStream && !m.HasOutputStream:
		w.Writef(") error")
	case !m.HasInput && !m.HasOutput && !m.HasInputStream && m.HasOutputStream:
		w.Writef("%s,) error", OutStreamer(m.OutputStreamType, resolver))
	case !m.HasInput && !m.HasOutput && m.HasInputStream && !m.HasOutputStream:
		w.Writef("%s, ) error", InStreamer(m.InputStreamType, resolver))
	case !m.HasInput && !m.HasOutput && m.HasInputStream && m.HasOutputStream:
		w.Writef("%s,) error", InOutStreamer(m.InputStreamType, m.OutputStreamType, resolver))
	case !m.HasInput && m.HasOutput && !m.HasInputStream && !m.HasOutputStream:
		w.Writef(") (")
		for _, o := range m.Output {
			w.Writef("%s,", MaybePointer(o, resolver))
		}
		w.Writef("error)")
	case !m.HasInput && m.HasOutput && !m.HasInputStream && m.HasOutputStream:
		w.Writef("*%s) error", m.ResponderName())
	case !m.HasInput && m.HasOutput && m.HasInputStream && !m.HasOutputStream:
		w.Writef("%s,) (", InStreamer(m.InputStreamType, resolver))
		for _, o := range m.Output {
			w.Writef("%s,", MaybePointer(o, resolver))
		}
		w.Writef("error)")
	case !m.HasInput && m.HasOutput && m.HasInputStream && m.HasOutputStream:
		w.Writef("*%s) error", m.ResponderName())
	case m.HasInput && !m.HasOutput && !m.HasInputStream && !m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s, ", MaybePointer(i.Type, resolver))
		}
		w.Writef(") error")
	case m.HasInput && !m.HasOutput && !m.HasInputStream && m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s,", MaybePointer(i.Type, resolver))
		}
		if isNamed {
			w.Writef("outStream ")
		}
		w.Writef("%s) error", OutStreamer(m.OutputStreamType, resolver))
	case m.HasInput && !m.HasOutput && m.HasInputStream && !m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s,", MaybePointer(i.Type, resolver))
		}
		if isNamed {
			w.Writef("inStream ")
		}
		w.Writef("%s) error", InStreamer(m.InputStreamType, resolver))
	case m.HasInput && !m.HasOutput && m.HasInputStream && m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s,", MaybePointer(i.Type, resolver))
		}
		if isNamed {
			w.Writef("inOutStream ")
		}
		w.Writef("%s) error", InOutStreamer(m.InputStreamType, m.OutputStreamType, resolver))
	case m.HasInput && m.HasOutput && !m.HasInputStream && !m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s,", MaybePointer(i.Type, resolver))
		}
		w.Writef(")(")
		for _, o := range m.Output {
			w.Writef("%s,", MaybePointer(o, resolver))
		}
		w.Writef("error)")
	case m.HasInput && m.HasOutput && !m.HasInputStream && m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s, ", MaybePointer(i.Type, resolver))
		}
		if isNamed {
			w.Writef("responder ")
		}
		w.Writef("*%s) error", m.ResponderName())
	case m.HasInput && m.HasOutput && m.HasInputStream && !m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s %s,", i.Name, MaybePointer(i.Type, resolver))
			}
		}
		w.Writef("*%s) error", m.ResponderName())
	case m.HasInput && m.HasOutput && m.HasInputStream && m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s, ", MaybePointer(i.Type, resolver))
		}
		if isNamed {
			w.Writef("responder ")
		}
		w.Writef("*%s) error", m.ResponderName())
	}
	w.Writelnf("")
}

func (m *MethodDefinition) HasResponder() bool {
	switch {
	case !m.HasInput && m.HasOutput && !m.HasInputStream && m.HasOutputStream:
		return true
	case !m.HasInput && m.HasOutput && m.HasInputStream && m.HasOutputStream:
		return true
	case m.HasInput && m.HasOutput && !m.HasInputStream && m.HasOutputStream:
		return true
	case m.HasInput && m.HasOutput && m.HasInputStream && !m.HasOutputStream:
		return true
	case m.HasInput && m.HasOutput && m.HasInputStream && m.HasOutputStream:
		return true
	default:
		return false
	}
}

func (m *MethodDefinition) ForClient() *MethodDefinition {
	return &MethodDefinition{
		Name:             m.Name,
		HasInput:         m.HasOutput,
		HasOutput:        m.HasInput,
		HasInputStream:   m.HasOutputStream,
		HasOutputStream:  m.HasInputStream,
		InputStreamType:  m.OutputStreamType,
		OutputStreamType: m.InputStreamType,
		Tree:             m.Tree,
	}
}
