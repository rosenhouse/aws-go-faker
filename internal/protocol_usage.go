package main

import (
	"bytes"
	"fmt"
	"go/build"
	"sort"
	"strings"

	"golang.org/x/tools/refactor/importgraph"
)

const ServicePackage = "github.com/aws/aws-sdk-go/service"
const ProtocolPackage = "github.com/aws/aws-sdk-go/private/protocol"

func isImmediateSubPackage(pkg, parent string) bool {
	prefix := parent + "/"
	if !strings.HasPrefix(pkg, prefix) {
		return false
	}

	shortName := pkg[len(prefix):]
	if strings.Contains(shortName, "/") {
		return false
	}

	return true
}

type ProtocolUsage map[string][]string

func GetUsage() ProtocolUsage {
	_, reverse, _ := importgraph.Build(&build.Default)

	protocolDependencies := map[string][]string{}
	for pkg, importers := range reverse {
		if isImmediateSubPackage(pkg, ProtocolPackage) {
			services := []string{}
			for importer, _ := range importers {
				if isImmediateSubPackage(importer, ServicePackage) {
					services = append(services, importer)
				}
			}
			if len(services) > 0 {
				sort.Strings(services)
				protocolDependencies[pkg] = services
			}
		}
	}

	return ProtocolUsage(protocolDependencies)
}

func (u ProtocolUsage) ForEach(action func(protocol string, dependentServices []string)) {
	sortedProtocols := make([]string, 0, len(u))
	for protocol := range u {
		sortedProtocols = append(sortedProtocols, protocol)
	}
	sort.Strings(sortedProtocols)
	for _, protocol := range sortedProtocols {
		action(protocol, u[protocol])
	}
}

func (u ProtocolUsage) AsMarkdown() string {
	buffer := &bytes.Buffer{}
	u.ForEach(func(protocol string, dependentServices []string) {
		toJoin := []string{}
		for _, service := range dependentServices {
			shortName := strings.TrimPrefix(service, ServicePackage+"/")
			toJoin = append(toJoin, fmt.Sprintf("`%s`", shortName))
		}

		fmt.Fprintf(buffer,
			"\n- [ ] *%s*: %s\n",
			strings.TrimPrefix(protocol, ProtocolPackage+"/"),
			strings.Join(toJoin, ", "),
		)
	})
	return buffer.String()
}

func main() {
	u := GetUsage()

	markdown := u.AsMarkdown()

	fmt.Println(markdown)
}
