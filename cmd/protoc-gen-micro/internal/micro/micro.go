package micro

import (
	"github.com/pthethanh/micro/cmd/protoc-gen-micro/internal/generator"
	pb "google.golang.org/protobuf/types/descriptorpb"
)

func init() {
	generator.RegisterPlugin(new(micro))
}

// micro is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for go-micro support.
type micro struct {
	gen *generator.Generator
}

// Name returns the name of this plugin, "micro".
func (g *micro) Name() string {
	return "micro"
}

// Init initializes the plugin.
func (g *micro) Init(gen *generator.Generator) {
	g.gen = gen
}

// P forwards to g.gen.P.
func (g *micro) P(args ...interface{}) { g.gen.P(args...) }

// Generate generates code for the services in the given file.
func (g *micro) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
	}
}

// GenerateImports generates the import declaration for this file.
func (g *micro) GenerateImports(file *generator.FileDescriptor, imports map[generator.GoImportPath]generator.GoPackageName) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	g.P("import (")
	g.P("grpc ", `"google.golang.org/grpc"`)
	if g.gen.GenGW {
		g.P(`"context"`)
		g.P("grpc ", `"google.golang.org/grpc"`)
		g.P(`"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"`)
	}
	g.P(")")
	g.P()
}

// generateService generates all the code for the named service.
func (g *micro) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {

	origServiceName := service.GetName()

	serviceName := generator.CamelCase(origServiceName)
	serviceAlias := "Unimplemented" + serviceName + "Server"

	g.P("func (", serviceAlias, ") ServiceDesc() *grpc.ServiceDesc{")
	g.P("return &", serviceName, "_ServiceDesc")
	g.P("}")

	if g.gen.GenGW {
		g.P()
		g.P("func (", serviceAlias, ") RegisterWithEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption){")
		g.P("Register" + serviceName + "HandlerFromEndpoint(ctx, mux, endpoint, opts)")
		g.P("}")
	}
}
