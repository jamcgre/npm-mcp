package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"errors"
	"sort"
)

type packageManifestArgs struct {
	Path string `json:"path"`
}

type packageInfoArgs struct {
	Name string `json:"name"`
}

type projectManifestArgs struct {
	Directory string `json:"directory"`
}

type packageLicense struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type licenseObject struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type packageFile struct {
	Dependencies     map[string]string `json:"dependencies"`
	DevDependencies  map[string]string `json:"devDependencies"`
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	Description      string            `json:"description"`
	Private          bool              `json:"private"`
	Scripts          map[string]string `json:"scripts"`
	PeerDependencies map[string]string `json:"peerDependencies"`
	Workspaces       []string          `json:"workspaces"`
	License          packageLicense    `json:"license"`
}

type registryPackage struct {
	Version string `json:"version"`
	License string `json:"license"`
}

type projectPackage struct {
	Path string `json:"path"`
	Manifest packageManifestResult `json:"manifest"`
}

type packageManifestResult struct {
	Dependencies     map[string]string `json:"dependencies"`
	DevDependencies  map[string]string `json:"devDependencies"`
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	Description      string            `json:"description"`
	Private          bool              `json:"private"`
	Scripts          map[string]string `json:"scripts"`
	PeerDependencies map[string]string `json:"peerDependencies"`
	Workspaces       []string          `json:"workspaces"`
	License          packageLicense    `json:"license"`
}

type PackageInfoResult struct {
	Version string `json:"version"`
	License string `json:"license"`
}

type projectManifestResult struct {
	Packages []projectPackage `json:"packages"`
}

type dependencyFreshnessArgs struct {
	Directory string `json:"directory"`
}

type freshnessRow struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Source string `json:"source"`
	Range string `json:"range"`
	Latest string `json:"latest"`
	Status string `json:"status"`
}

type dependencyFreshnessResult struct {
	Rows []freshnessRow `json:"rows"`
}

func (license *packageLicense) UnmarshalJSON(data []byte) error {
	var licenseType string
	if err := json.Unmarshal(data, &licenseType); err == nil {
		license.Type = licenseType
		return nil
	}

	var licenseData licenseObject
	if err := json.Unmarshal(data, &licenseData); err == nil {
		license.Type = licenseData.Type
		license.URL = licenseData.URL
		return nil
	}

	return fmt.Errorf("invalid license format")
}

func packageManifest(ctx context.Context, req *mcp.CallToolRequest, args packageManifestArgs) (*mcp.CallToolResult, packageManifestResult, error) {
	path, err := os.ReadFile(args.Path)
	if err != nil {
		return nil, packageManifestResult{}, err
	}

	var file packageFile
	err = json.Unmarshal(path, &file)
	if err != nil {
		return nil, packageManifestResult{}, err
	}

	if file.Dependencies == nil {
		file.Dependencies = make(map[string]string)
	}
	if file.DevDependencies == nil {
		file.DevDependencies = make(map[string]string)
	}
	if file.PeerDependencies == nil {
		file.PeerDependencies = make(map[string]string)
	}
	if file.Scripts == nil {
		file.Scripts = make(map[string]string)
	}
	if file.Workspaces == nil {
		file.Workspaces = make([]string, 0)
	}

	return nil, packageManifestResult{Dependencies: file.Dependencies, DevDependencies: file.DevDependencies, Name: file.Name, Version: file.Version, Description: file.Description, Private: file.Private, Scripts: file.Scripts, PeerDependencies: file.PeerDependencies, Workspaces: file.Workspaces, License: file.License}, nil
}

func projectManifest(ctx context.Context, req *mcp.CallToolRequest, args projectManifestArgs) (*mcp.CallToolResult, projectManifestResult, error) {
	_, packageManifests, err := packageManifest(ctx, req, packageManifestArgs{Path: filepath.Join(args.Directory, "package.json")})

	if err != nil {
		return nil, projectManifestResult{}, err
	}

	packages := []projectPackage{}

	packages = append(packages, projectPackage{Path: filepath.Join(args.Directory, "package.json"), Manifest: packageManifests})

	for _, workspace := range packageManifests.Workspaces {
		matches, err := filepath.Glob(filepath.Join(args.Directory, workspace))
		if err != nil {
			return nil, projectManifestResult{}, err
		}

		for _, match := range matches {
			packagePath := filepath.Join(match, "package.json")
			if _, err := os.Stat(packagePath); err != nil {
				continue
			}

			_, workspaceManifest, err := packageManifest(ctx, req, packageManifestArgs{Path: packagePath})
			if err != nil {
				return nil, projectManifestResult{}, err
			}
			packages = append(packages, projectPackage{Path: packagePath, Manifest: workspaceManifest})
		}
	}

	return nil, projectManifestResult{Packages: packages}, nil
}


func packageInfo(ctx context.Context, req *mcp.CallToolRequest, args packageInfoArgs) (*mcp.CallToolResult, PackageInfoResult, error) {
	url := "https://registry.npmjs.org/" + args.Name + "/latest"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, PackageInfoResult{}, err
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, PackageInfoResult{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, PackageInfoResult{}, fmt.Errorf("registry returned %s", resp.Status)
	}

	var doc registryPackage
	err = json.NewDecoder(resp.Body).Decode(&doc)
	if err != nil {
		return nil, PackageInfoResult{}, err
	}

	return nil, PackageInfoResult{Version: doc.Version, License: doc.License}, nil
}

var errorPackageMissing = errors.New("package not found")

var lookupLatest = func(ctx context.Context, name string) (string, error) {
	url := "https://registry.npmjs.org/" + name + "/latest"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", errorPackageMissing
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registry returned %s", resp.Status)
	}

   var doc struct {
		Version string `json:"version"`
	}
	err = json.NewDecoder(resp.Body).Decode(&doc)
	if err != nil {
		return "", err
	}
	return doc.Version, nil
}

func dependencyFreshness(ctx context.Context, req *mcp.CallToolRequest, args dependencyFreshnessArgs) (*mcp.CallToolResult, dependencyFreshnessResult, error) {
	_, manifests, err := projectManifest(ctx, req, projectManifestArgs{Directory: args.Directory})
	if err != nil {
		return nil, dependencyFreshnessResult{}, err
	}

	rows := []freshnessRow{}

	for _, pkg := range manifests.Packages {
		for name, rng := range pkg.Manifest.Dependencies {
			row := freshnessRow{
				Path: pkg.Path,
				Name: name,
				Source: "dependencies",
				Range: rng,
			}
			if !supportedRange(rng) {
				row.Status = "unparsed"
			} else if rng == "*" {
				row.Status = "inside"
			} else {
				latest, err := lookupLatest(ctx, name)
				if errors.Is(err, errorPackageMissing) {
					row.Status = "missing"
				} else if err != nil {
					return nil, dependencyFreshnessResult{}, err
				} else {
					row.Latest = latest
					row.Status = rangeStatus(rng, latest)
				}
			}
			rows = append(rows, row)
		}
		for name, rng := range pkg.Manifest.DevDependencies {
			row := freshnessRow{
				Path: pkg.Path,
				Name: name,
				Source: "devDependencies",
				Range: rng,
			}
			if !supportedRange(rng) {
				row.Status = "unparsed"
			} else if rng == "*" {
				row.Status = "inside"
			} else {
				latest, err := lookupLatest(ctx, name)
				if errors.Is(err, errorPackageMissing) {
					row.Status = "missing"
				} else if err != nil {
					return nil, dependencyFreshnessResult{}, err
				} else {
					row.Latest = latest
					row.Status = rangeStatus(rng, latest)
				}
			}
			rows = append(rows, row)
		}
		for name, rng := range pkg.Manifest.PeerDependencies {
			row := freshnessRow{
				Path: pkg.Path,
				Name: name,
				Source: "peerDependencies",
				Range: rng,
			}
			if !supportedRange(rng) {
				row.Status = "unparsed"
			} else if rng == "*" {
				row.Status = "inside"
			} else {
				latest, err := lookupLatest(ctx, name)
				if errors.Is(err, errorPackageMissing) {
					row.Status = "missing"
				} else if err != nil {
					return nil, dependencyFreshnessResult{}, err
				} else {
					row.Latest = latest
					row.Status = rangeStatus(rng, latest)
				}
			}
			rows = append(rows, row)
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Path != rows[j].Path {
			return rows[i].Path < rows[j].Path
		}
		if rows[i].Source != rows[j].Source {
			return rows[i].Source < rows[j].Source
		}
		return rows[i].Name < rows[j].Name
	})

	return nil, dependencyFreshnessResult{Rows: rows}, nil
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "npm-mcp", Version: "1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "package_manifest",
		Description: "Read a package.json and return the manifest",
	}, packageManifest)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "package_info",
		Description: "Return version and license of package from npm registry",
	}, packageInfo)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_manifest",
		Description: "Read a project's package.json and package.json that match the workspace pattern",
	}, projectManifest)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dependency_freshness",
		Description: "Report whether the latest registry version of each declared dependency falls outside that dependency's range",
	}, dependencyFreshness)

	err := server.Run(context.Background(), &mcp.StdioTransport{})
	if err != nil {
		log.Fatal(err)
	}
}
