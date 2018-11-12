package main

import (
	"go/build"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/spf13/afero"

	"github.com/izumin5210/grapi/pkg/cli"
	"github.com/izumin5210/grapi/pkg/gencmd"
	gencmdtesting "github.com/izumin5210/grapi/pkg/gencmd/testing"
	"github.com/izumin5210/grapi/pkg/grapicmd"
	"github.com/izumin5210/grapi/pkg/grapicmd/util/fs"
)

func TestType(t *testing.T) {
	cases := []struct {
		test  string
		args  []string
		files []string
	}{
		{
			test:  "simple",
			args:  []string{"book"},
			files: []string{"api/protos/type/book.proto"},
		},
		{
			test:  "nested",
			args:  []string{"foo/user"},
			files: []string{"api/protos/type/foo/user.proto"},
		},
		{
			test:  "camel",
			args:  []string{"foo/barUser"},
			files: []string{"api/protos/type/foo/bar_user.proto"},
		},
		{
			test:  "snake",
			args:  []string{"foo/bar_user"},
			files: []string{"api/protos/type/foo/bar_user.proto"},
		},
		{
			test:  "kebab",
			args:  []string{"foo/bar-user"},
			files: []string{"api/protos/type/foo/bar_user.proto"},
		},
	}

	defer func(c build.Context) { fs.BuildContext = c }(fs.BuildContext)
	fs.BuildContext = build.Context{GOPATH: "/go"}
	rootDir := cli.RootDir("/go/src/testapp")

	createGenApp := func(ctx *gencmd.Ctx, cmd *gencmd.Command) (*gencmd.App, error) {
		return gencmdtesting.NewTestApp(ctx, cmd, cli.NopUI)
	}
	createCmd := func(t *testing.T, fs afero.Fs) gencmd.Executor {
		ctx := &grapicmd.Ctx{
			FS:      fs,
			RootDir: rootDir,
		}
		return buildCommand(gencmd.WithGrapiCtx(ctx), gencmd.WithCreateAppFunc(createGenApp))
	}

	for _, tc := range cases {
		t.Run(tc.test, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			afero.WriteFile(fs, rootDir.Join("grapi.toml"), []byte{}, 0755)

			t.Run("generate", func(t *testing.T) {
				cmd := createCmd(t, fs)
				cmd.Command().SetArgs(append([]string{"generate"}, tc.args...))
				err := cmd.Execute()

				if err != nil {
					t.Errorf("returned an error: %+v", err)
				}

				for _, file := range tc.files {
					t.Run(file, func(t *testing.T) {
						data, err := afero.ReadFile(fs, rootDir.Join(file))

						if err != nil {
							t.Errorf("returned an error: %v", err)
						}

						cupaloy.SnapshotT(t, string(data))
					})
				}
			})

			t.Run("destroy", func(t *testing.T) {
				cmd := createCmd(t, fs)
				cmd.Command().SetArgs(append([]string{"destroy"}, tc.args...))
				err := cmd.Execute()

				if err != nil {
					t.Errorf("returned an error: %+v", err)
				}

				for _, file := range tc.files {
					t.Run(file, func(t *testing.T) {
						ok, err := afero.Exists(fs, rootDir.Join(file))

						if err != nil {
							t.Errorf("Exists(fs, %q) returned an error: %v", file, err)
						}

						if ok {
							t.Errorf("%q should not exist", file)
						}
					})
				}
			})
		})
	}
}
