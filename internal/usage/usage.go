// Package usage provides a tool for dynamically determining
// which AWS services use which AWS protocols.
//
// It supports tests of the detect package, ensuring that the protocol map
// stays up to date
//
// It is not run during normal use of awsfaker because building up the
// requisite reverse import path takes several seconds to complete.
package usage

import (
	"sort"
	"strings"

	"golang.org/x/tools/refactor/importgraph"
)

const ServicePackage = "github.com/aws/aws-sdk-go/service"
const ProtocolPackage = "github.com/aws/aws-sdk-go/private/protocol"

func shortenProtocol(longProtocol string) string {
	return strings.TrimPrefix(longProtocol, ProtocolPackage+"/")
}

func shortenService(longService string) string {
	return strings.TrimPrefix(longService, ServicePackage+"/")
}

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

// Return the ProtocolUsage, expressed in short package names, given a reverse import graph
//
// To get the reverse import path, run:
//    _, reverseImports, _ := importgraph.Build(&build.Default)
func GetShortUsage(reverseImports importgraph.Graph) ProtocolUsage {
	shortUsage := map[string][]string{}
	usage := getUsage(reverseImports)
	for protocol, services := range usage {
		shortServices := make([]string, len(services))
		for i, service := range services {
			shortServices[i] = shortenService(service)
		}

		shortUsage[shortenProtocol(protocol)] = shortServices
	}
	return shortUsage
}

func getUsage(reverseImports importgraph.Graph) ProtocolUsage {
	protocolDependencies := map[string][]string{}
	for pkg, importers := range reverseImports {
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

	return protocolDependencies
}
