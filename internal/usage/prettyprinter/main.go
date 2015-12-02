package main

import (
	"bytes"
	"fmt"
	"go/build"
	"sort"
	"strings"

	"github.com/rosenhouse/awsfaker/internal/usage"

	"golang.org/x/tools/refactor/importgraph"
)

func ForEach(u usage.ProtocolUsage, action func(protocol string, dependentServices []string)) {
	sortedProtocols := make([]string, 0, len(u))
	for protocol := range u {
		sortedProtocols = append(sortedProtocols, protocol)
	}
	sort.Strings(sortedProtocols)
	for _, protocol := range sortedProtocols {
		action(protocol, u[protocol])
	}
}

func ToMarkdown(u usage.ProtocolUsage) string {
	buffer := &bytes.Buffer{}
	ForEach(u, func(protocol string, dependentServices []string) {
		toJoin := []string{}
		for _, service := range dependentServices {
			toJoin = append(toJoin, fmt.Sprintf("`%s`", service))
		}

		fmt.Fprintf(buffer, "\n- *%s*: %s\n", protocol, strings.Join(toJoin, ", "))
	})
	return buffer.String()
}

// Print a Markdown report of which AWS services use which AWS protocols
func main() {
	_, reverseImports, _ := importgraph.Build(&build.Default)
	u := usage.GetShortUsage(reverseImports)

	markdown := ToMarkdown(u)

	fmt.Println(markdown)
}
