package main

import (
	"context"
	"log"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"encoding/json"
	"os"
	"fmt"
	"net/http"
)

type packageDependenciesArgs struct {
	Path string `json:"path"`
}

type packageInfoArgs struct {
	Name string `json:"name"`
}

type packageFile struct {
	Dependencies 	map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type registryPackage struct {
	Version string `json:"version"`
	License string `json:"license"`
}

type packageDependenciesResult struct {
	Dependencies 	[]string `json:"dependencies"`
	DevDependencies []string `json:"devDependencies"`
}

type PackageInfoResult struct {
	Version string `json:"version"`
	License string `json:"license"`
}

func packageDependencies(ctx context.Context, req *mcp.CallToolRequest, args packageDependenciesArgs) (*mcp.CallToolResult, packageDependenciesResult, error) {
	path, err := os.ReadFile(args.Path)
	if (err != nil) {
		return nil, packageDependenciesResult{}, err
	}

	var file packageFile
	err = json.Unmarshal(path, &file)
	if (err != nil) {
		return nil, packageDependenciesResult{}, err
	}

	var dependencies []string

	for name, version := range file.Dependencies {
		dependencies = append(dependencies, name + "@" + version)
	}

	var devDependencies []string

	for name, version := range file.DevDependencies {
		devDependencies = append(devDependencies, name + "@" + version)
	}

	return nil, packageDependenciesResult{Dependencies: dependencies, DevDependencies: devDependencies}, nil
}

func packageInfo(ctx context.Context, req *mcp.CallToolRequest, args packageInfoArgs) (*mcp.CallToolResult, PackageInfoResult, error) {
	url := "https://registry.npmjs.org/" + args.Name + "/latest"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if (err != nil) {
		return nil, PackageInfoResult{}, err
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if (err != nil) {
		return nil, PackageInfoResult{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, PackageInfoResult{}, fmt.Errorf("registry returned %s", resp.Status)
	}

	var doc registryPackage
	err = json.NewDecoder(resp.Body).Decode(&doc)
	if (err != nil) {
		return nil, PackageInfoResult{}, err
	}

	return nil, PackageInfoResult{Version: doc.Version, License: doc.License}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "npm-mcp", Version: "1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name: 			"package_dependencies",
		Description: 	"Read a package.json and return dependency names and versions",
	}, packageDependencies)

	mcp.AddTool(server, &mcp.Tool{
		Name: 			"package_info",
		Description: 	"Return version and license of package from npm registry",
	}, packageInfo)

	err := server.Run(context.Background(), &mcp.StdioTransport{})
	if (err != nil) {
		log.Fatal(err)
	}
}