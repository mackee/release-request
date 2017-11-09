# Release

Estimated time of deploying: {{ .Call "deploy_timing" }}

## Merged

{{ range .MergedPullRequests -}}
- [x] #{{ .Number }} {{ .Title }} by @{{ .Author }}
{{ end }}
{{ with .OpenedPullRequests }}

## Pendings

{{ range . -}}
- [ ] #{{ .Number }} {{ .Title }} by @{{ .Author }}
{{ end }}
{{ end -}}

## Tasks

{{ with .CallIf "has_assets" -}}
- [ ] generate assets

```console
$ go generate
```

{{ end -}}

- [ ] bump tag

```
$ git tag v1.x.y
$ git push origin v1.x.y
```

- [ ] publish release

```
$ make releases
```
