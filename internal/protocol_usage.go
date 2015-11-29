package main

import (
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

func main() {
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

	sortedProtocols := make([]string, 0, len(protocolDependencies))
	for protocol := range protocolDependencies {
		sortedProtocols = append(sortedProtocols, protocol)
	}
	sort.Strings(sortedProtocols)
	for _, protocol := range sortedProtocols {
		dependentServices := protocolDependencies[protocol]
		fmt.Printf("%s:\n", strings.TrimPrefix(protocol, ProtocolPackage+"/"))
		for _, service := range dependentServices {
			fmt.Printf("\t%s\n", strings.TrimPrefix(service, ServicePackage+"/"))
		}
	}
}
