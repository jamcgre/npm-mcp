# npm-mcp

MCP server that reads npm `package.json` files and checks declared dependency ranges against the latest version on the npm registry.

The server speaks MCP over stdin and stdout. It has four tools.

## Tools

### `package_manifest`

Reads one `package.json`. Argument: `path`, the file path.

Returns `name`, `version`, `description`, `private`, `license`, `scripts`, `workspaces`, `dependencies`, `devDependencies`, and `peerDependencies`. A missing `private` is `false`. A missing map is `{}` and a missing `workspaces` list is `[]`. `license` may be a string or an object with `type` and `url`. Dependency maps are package name to range.

### `package_info`

Looks up one published package. Argument: `name`.

Sends `GET https://registry.npmjs.org/<name>/latest` and returns `version` and `license` as strings. A scoped name keeps its slash, as in `@types/react`. The license is read as a string, so a registry document whose license is an object fails this call.

### `project_manifest`

Reads a project. Argument: `directory`.

Reads `<directory>/package.json`, then expands each `workspaces` pattern with a glob. The root package is first. Each entry's `path` is the `package.json` file. A match with no `package.json` is skipped. `*` matches one path segment, so a literal name such as `client` works. Patterns that contain `**` are left unexpanded. Child `workspaces` fields are not followed.

### `dependency_freshness`

Checks every declared dependency in a project. Argument: `directory`.

Calls `project_manifest`, then walks `dependencies`, `devDependencies`, and `peerDependencies`. Each row has `path`, `name`, `source`, `range`, `latest`, and `status`. Rows are sorted by path, then source, then name.

`status` is one of:

| Status | Meaning |
| --- | --- |
| `inside` | The latest registry version satisfies the range. |
| `outside` | The latest registry version does not satisfy the range. |
| `unparsed` | The range shape is not compared. `latest` is empty and the registry is not called. |
| `missing` | The registry returned 404. `latest` is empty. |

`*` is `inside` and does not call the registry. An exact range is three numbers, such as `1.2.3`. A caret (`^`) allows versions up to, but not including, npm's next breaking boundary: the next major for `^1`, `^1.2`, `^1.2.3`, and `^0`; the next minor for `^0.2` and `^0.2.3`; the next patch for `^0.0.3`. A tilde (`~`) with one part allows the next major. A tilde with two or three parts allows the next minor. The lower bound is inclusive and the upper bound is exclusive.

Anything else is `unparsed`, including `||`, spaces, comparison operators, prerelease versions, `x` ranges, and git, file, or URL specs. A registry failure other than 404 fails the whole call.

`^19.1.1` against `19.3.1` is `inside`. `^4.21.2` against `5.3.0` is `outside`.

## Run

Install Go 1.27 or newer, then build from this directory:

```sh
go build
```

Point an MCP client at the binary. In Cursor, a project `.cursor/mcp.json` looks like this:

```json
{
  "mcpServers": {
    "npm-mcp": {
      "command": "/absolute/path/to/npm-mcp"
    }
  }
}
```

On Windows the binary is `npm-mcp.exe`. Use an absolute path. Restart the MCP client after rebuilding so it picks up the new binary.

```sh
go test
```
