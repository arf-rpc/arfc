package common

import (
	"fmt"
	"github.com/arf-rpc/arfc/arf/strcase"
	"github.com/arf-rpc/idl/ast"
	"slices"
	"strings"
)

func MaybePointer(t ast.Type) string {
	str := ConvertType(t)
	if IsUserType(t) {
		str = "*" + str
	}
	return str
}

func PackageName(t ast.Object) string {
	comps := strings.Split(t.BaseFQN(), ".")
	return comps[len(comps)-1]
}

func ConvertType(t ast.Type) string {
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
		return "*" + ConvertType(v.Type)
	case *ast.ArrayType:
		return "[]" + ConvertType(v.Type)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", ConvertType(v.Key), ConvertType(v.Value))
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
		switch a := v.ResolvedType.(type) {
		case *ast.Enum:
			return EnumName(a)
		case *ast.Struct:
			return StructName(a)
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

func InStreamer(t ast.Type) string {
	return fmt.Sprintf("arf.InStreamer[%s]", MaybePointer(t))
}
func OutStreamer(t ast.Type) string {
	return fmt.Sprintf("arf.OutStreamer[%s]", MaybePointer(t))
}
func InOutStreamer(i, o ast.Type) string {
	return fmt.Sprintf("arf.InOutStreamer[%s, %s]", MaybePointer(i), MaybePointer(o))
}

func (m *MethodDefinition) ResponderName() string {
	return fmt.Sprintf("%s%sResponder", strcase.ToCamel(PackageName(m.Method)), strcase.ToCamel(m.Name))
}

func (m *MethodDefinition) BuildInterfaceSignature(w *Writer) {
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
		w.Writef("%s,) error", OutStreamer(m.OutputStreamType))
	case !m.HasInput && !m.HasOutput && m.HasInputStream && !m.HasOutputStream:
		w.Writef("%s, ) error", InStreamer(m.InputStreamType))
	case !m.HasInput && !m.HasOutput && m.HasInputStream && m.HasOutputStream:
		w.Writef("%s,) error", InOutStreamer(m.InputStreamType, m.OutputStreamType))
	case !m.HasInput && m.HasOutput && !m.HasInputStream && !m.HasOutputStream:
		w.Writef(") (")
		for _, o := range m.Output {
			w.Writef("%s,", MaybePointer(o))
		}
		w.Writef("error)")
	case !m.HasInput && m.HasOutput && !m.HasInputStream && m.HasOutputStream:
		w.Writef("*%s) error", m.ResponderName())
	case !m.HasInput && m.HasOutput && m.HasInputStream && !m.HasOutputStream:
		w.Writef("%s,) (", InStreamer(m.InputStreamType))
		for _, o := range m.Output {
			w.Writef("%s,", MaybePointer(o))
		}
		w.Writef("error)")
	case !m.HasInput && m.HasOutput && m.HasInputStream && m.HasOutputStream:
		w.Writef("*%s) error", m.ResponderName())
	case m.HasInput && !m.HasOutput && !m.HasInputStream && !m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s, ", MaybePointer(i.Type))
		}
		w.Writef(") error")
	case m.HasInput && !m.HasOutput && !m.HasInputStream && m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s,", MaybePointer(i.Type))
		}
		if isNamed {
			w.Writef("outStream ")
		}
		w.Writef("%s) error", OutStreamer(m.OutputStreamType))
	case m.HasInput && !m.HasOutput && m.HasInputStream && !m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s,", MaybePointer(i.Type))
		}
		if isNamed {
			w.Writef("inStream ")
		}
		w.Writef("%s) error", InStreamer(m.InputStreamType))
	case m.HasInput && !m.HasOutput && m.HasInputStream && m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s,", MaybePointer(i.Type))
		}
		if isNamed {
			w.Writef("inOutStream ")
		}
		w.Writef("%s) error", InOutStreamer(m.InputStreamType, m.OutputStreamType))
	case m.HasInput && m.HasOutput && !m.HasInputStream && !m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s,", MaybePointer(i.Type))
		}
		w.Writef(")(")
		for _, o := range m.Output {
			w.Writef("%s,", MaybePointer(o))
		}
		w.Writef("error)")
	case m.HasInput && m.HasOutput && !m.HasInputStream && m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s, ", MaybePointer(i.Type))
		}
		if isNamed {
			w.Writef("responder ")
		}
		w.Writef("*%s) error", m.ResponderName())
	case m.HasInput && m.HasOutput && m.HasInputStream && !m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s %s,", i.Name, MaybePointer(i.Type))
			}
		}
		w.Writef("*%s) error", m.ResponderName())
	case m.HasInput && m.HasOutput && m.HasInputStream && m.HasOutputStream:
		for _, i := range m.Inputs {
			if isNamed {
				w.Writef("%s ", i.Name)
			}
			w.Writef("%s, ", MaybePointer(i.Type))
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
