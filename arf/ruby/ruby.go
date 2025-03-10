package ruby

import (
	"fmt"
	"github.com/arf-rpc/arfc/arf/common"
	"github.com/arf-rpc/arfc/arf/strcase"
	"github.com/arf-rpc/arfc/output"
	"github.com/arf-rpc/idl/ast"
	"github.com/urfave/cli/v2"
	"path/filepath"
	"regexp"
	"strings"
)

func NewGenerator(tree *ast.PackageTree) common.Generator {
	return &Generator{
		t: tree,
		w: &common.Writer{},
	}
}

type Generator struct {
	t *ast.PackageTree
	w *common.Writer
}

func (g *Generator) GenFile(ctx *cli.Context) (data []byte, targetDir string, targetFile string) {
	moduleMapping := map[string]string{}
	for _, mod := range ctx.StringSlice("ruby-module") {
		comps := strings.SplitN(mod, "=", 2)
		if len(comps) != 2 {
			output.Errorf("Invalid value for ruby-module: %s", mod)
		}
		moduleMapping[strings.TrimSpace(comps[0])] = strings.TrimSpace(comps[1])
	}

	mods := strings.Split(g.t.Package, ".")
	if mod, ok := moduleMapping[g.t.Package]; ok {
		modValidator := regexp.MustCompile(`^([A-Z][a-zA-Z0-9]*)(::([A-Z][a-zA-Z0-9]*))*$`)
		if !modValidator.MatchString(mod) {
			output.Errorf("Invalid module %s for ruby-module: %s", g.t.Package, mod)
		}
		mods = strings.Split(mod, "::")
	}

	isFlat := ctx.Bool("ruby-flat")
	if isFlat {
		targetDir = ctx.String("output")
	} else {
		dirs := make([]string, 0, len(mods)-1)
		for _, mod := range mods[:len(mods)-1] {
			dirs = append(dirs, strcase.ToSnake(mod))
		}
		targetDir = filepath.Join(append([]string{ctx.String("output")}, dirs...)...)
	}

	targetFile = strcase.ToSnake(mods[len(mods)-1]) + ".arf.rb"

	g.writeHeader()
	for _, mod := range mods {
		g.w.Writef("module %s\n", strcase.ToCamel(mod))
		g.w.IncreaseIndent()
	}

	for _, s := range g.t.Structures {
		g.makeStruct(&s)
	}

	for _, s := range g.t.Services {
		g.makeService(&s)
	}

	for _, s := range g.t.Services {
		g.makeClient(&s)
	}

	for range mods {
		g.w.DecreaseIndent()
		g.w.Writef("end\n")
	}

	data = []byte(g.w.String())
	return
}

func (g *Generator) makeStruct(s *ast.Struct) {
	g.writeComments(s.Comment)
	if ann := s.Annotations.ByName("deprecated"); ann != nil {
		g.w.Writelnf("# Deprecated: %s", ann.Arguments[0])
	}
	g.w.Writelnf("class %s < Arf::RPC::Struct", strcase.ToCamel(s.Name))
	g.w.IncreaseIndent()

	g.w.Writelnf("arf_struct_id \"%s\"", common.CanonicalStructName(g.t.Package, s))

	for _, f := range s.Fields {
		g.w.Writelnf("field %d, :%s, %s",
			f.ID,
			strcase.ToSnake(f.Name),
			ConvertType(f.Type))
	}

	for _, st := range s.Structs {
		g.makeStruct(&st)
	}

	for _, e := range s.Enums {
		g.makeEnum(&e)
	}

	g.w.DecreaseIndent()
	g.w.Writelnf("end")
}

func (g *Generator) makeEnum(e *ast.Enum) {
	g.writeComments(e.Comment)
	g.w.Writelnf("class %s < Arf::RPC::Enum", strcase.ToCamel(e.Name))
	g.w.IncreaseIndent()
	for _, v := range e.Members {
		g.w.Writelnf("option %s: %d", strcase.ToSnake(v.Name), v.Value)
	}
	g.w.DecreaseIndent()
	g.w.Writelnf("end")
}

func (g *Generator) writeComments(c []string) {
	g.w.Break()
	for _, c := range c {
		g.w.Writelnf("#%s", c)
	}
}

func (g *Generator) writeHeader() {
	g.w.Writelnf("# Code generated by arfc. DO NOT EDIT.\n")
}

func (g *Generator) makeService(s *ast.Service) {
	g.writeComments(s.Comment)
	g.w.Writelnf("class %s < Arf::RPC::ServiceBase", strcase.ToCamel(s.Name))
	g.w.IncreaseIndent()
	g.w.Writelnf("arf_service_id \"%s/%s\"", g.t.Package, strcase.ToSnake(s.Name))
	for _, method := range s.Methods {
		g.makeRPCDefinition(method)
	}
	g.w.DecreaseIndent()
	g.w.Writelnf("end")
}

func (g *Generator) makeRPCDefinition(method *ast.ServiceMethod) {
	g.w.Writef("rpc :%s", strcase.ToSnake(method.Name))
	g.w.IncreaseIndent()

	if len(method.Params) > 0 {
		g.w.Writelnf(",")
		var allInputs []string
		var stream ast.Type
		for _, i := range method.Params {
			if i.Stream {
				stream = i.Type
			} else {
				allInputs = append(allInputs, fmt.Sprintf("%s: %s", *i.Name, ConvertType(i.Type)))
			}
		}
		if stream != nil {
			allInputs = append(allInputs, fmt.Sprintf("_stream: InputStream[%s]", ConvertType(stream)))
		}
		g.w.Writef("inputs: { %s }", strings.Join(allInputs, ", "))
	}

	if len(method.Returns) > 0 {
		g.w.Writelnf(",")
		var allOutputs []string
		var stream ast.Type
		for _, i := range method.Returns {
			if i.Stream {
				stream = i.Type
			} else {
				allOutputs = append(allOutputs, fmt.Sprintf("%s", ConvertType(i.Type)))
			}
		}
		if stream != nil {
			allOutputs = append(allOutputs, fmt.Sprintf("OutputStream[%s]", ConvertType(stream)))
		}
		g.w.Writef("outputs: [%s]", strings.Join(allOutputs, ", "))
	}

	g.w.DecreaseIndent()
	g.w.Break()
}

func (g *Generator) makeClient(s *ast.Service) {
	g.writeComments(s.Comment)
	g.w.Writelnf("class %sClient < Arf::RPC::ClientBase", strcase.ToCamel(s.Name))
	g.w.IncreaseIndent()
	g.w.Writelnf("arf_service_id \"%s/%s\"", g.t.Package, strcase.ToSnake(s.Name))
	for _, method := range s.Methods {
		g.makeRPCDefinition(method)
	}
	g.w.DecreaseIndent()
	g.w.Writelnf("end")
}

func ConvertType(t ast.Type) string {
	switch v := t.(type) {
	case *ast.PrimitiveType:
		return ":" + v.Name
	case *ast.OptionalType:
		return ConvertType(v.Type) + ", optional: true"
	case *ast.ArrayType:
		t := ConvertType(v.Type)
		if strings.HasPrefix(t, ":") {
			return "ArrayType[" + t + "]"
		} else {
			return "ArrayType[" + t + "].bind(self)"
		}
	case *ast.MapType:
		key, value := ConvertType(v.Key), ConvertType(v.Value)
		if strings.HasSuffix(key, ":") && strings.HasPrefix(value, ":") {
			return fmt.Sprintf("MapType[%s, %s]", key, value)
		} else {
			return fmt.Sprintf("MapType[%s, %s].bind(self)", key, value)
		}
	case *ast.SimpleUserType:
		switch a := v.ResolvedType.(type) {
		case *ast.Enum:
			return fmt.Sprintf("\"%s\"", a.Name)
		case *ast.Struct:
			return fmt.Sprintf("\"%s\"", a.Name)
		default:
			return "INVALID"
		}
	case *ast.FullQualifiedType:
		switch a := v.ResolvedType.(type) {
		case *ast.Enum:
			return fmt.Sprintf("\"%s\"", a.Name)
		case *ast.Struct:
			return fmt.Sprintf("\"%s\"", a.Name)
		default:
			return "INVALID"
		}
	default:
		return "INVALID"
	}
}
