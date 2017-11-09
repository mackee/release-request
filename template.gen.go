package main

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var _Assetsf000cb359297bc9533188180f9bc040ef1fdcb3e = "title: '[DEPLOY] {{ .Call \"deploy_title\" }}'\nrepo:\n  owner: mackee\n  repo: release-request\n  base: master\n  head: develop\nlabels:\n  - deploy\ntarget_label: deploy ready\nmarkdown: release.md\ncalls:\n  - name: deploy_timing\n    command: |\n      perl -MTime::Piece -E '\n      my $now = Time::Piece->new;\n      for my $hms (qw/16:30:00 19:45:00 23:45:00/) {\n          next if $now->hms gt $hms;\n          say $now->ymd . \" \" . $hms;\n          last;\n      }'\n  - name: deploy_title\n    command: date +\"%Y-%m-%d %H:%M\"\n  - name: has_assets\n    command: |\n      perl -MJSON -E '\n      my $json = do { local $/; <STDIN> };\n      my $releaser = JSON::decode_json($json);\n      my $files = $releaser->{comparison}->{files};\n      my @ddls = grep { $_->{name} =~ m!^_example/.*$! } @$files;\n      @ddls > 0 ? exit 0 : exit 1;\n      '\ntoken: \"<your token>\"\n"
var _Assets9d7da04f478156ce3962423189e531c0785612c3 = "# Release\n\nEstimated time of deploying: {{ .Call \"deploy_timing\" }}\n\n## Merged\n\n{{ range .MergedPullRequests -}}\n- [x] #{{ .Number }} {{ .Title }} by @{{ .Author }}\n{{ end -}}\n{{ with .OpenedPullRequests }}\n\n## Pendings\n\n{{ range . -}}\n- [ ] #{{ .Number }} {{ .Title }} by @{{ .Author }}\n{{ end -}}\n\n## Tasks\n\n{{ with .CallIf \"has_assets\" -}}\n- [ ] generate assets\n\n```console\n$ go generate\n```\n\n{{ end -}}\n\n- [ ] bump tag\n\n```\n$ git tag v1.x.y\n$ git push origin v1.x.y\n```\n\n- [ ] publish release\n\n```\n$ make releases\n```\n"

// Assets returns go-assets FileSystem
var Assets = assets.NewFileSystem(map[string][]string{}, map[string]*assets.File{
	"config.yml": &assets.File{
		Path:     "config.yml",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1510208122, 1510208122000000000),
		Data:     []byte(_Assetsf000cb359297bc9533188180f9bc040ef1fdcb3e),
	}, "release.md": &assets.File{
		Path:     "release.md",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1510208314, 1510208314000000000),
		Data:     []byte(_Assets9d7da04f478156ce3962423189e531c0785612c3),
	}}, "")
