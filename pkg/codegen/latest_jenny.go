package codegen

import (
	"fmt"
	"path/filepath"

	"github.com/grafana/codejen"
)

// LatestJenny returns a jenny that runs another jenny for only the latest
// schema in a DeclForGen, and prefixes the resulting file with the provided
// parentdir (e.g. "pkg/kinds/") and with a directory based on the kind's
// machine name (e.g. "dashboard/").
func LatestJenny(parentdir string, inner codejen.OneToOne[SchemaForGen]) OneToOne {
	if inner == nil {
		panic("inner jenny must not be nil")
	}

	return &latestj{
		parentdir: parentdir,
		inner:     inner,
	}
}

type latestj struct {
	parentdir string
	inner     codejen.OneToOne[SchemaForGen]
}

func (j *latestj) JennyName() string {
	return "LatestJenny"
}

func (j *latestj) Generate(decl *DeclForGen) (*codejen.File, error) {
	if decl.IsRaw() {
		return nil, nil
	}
	comm := decl.Properties.Common()
	sfg := SchemaForGen{
		Name:    comm.Name,
		Schema:  decl.Lineage().Latest(),
		IsGroup: comm.LineageIsGroup,
	}

	f, err := j.inner.Generate(sfg)
	if err != nil {
		return nil, fmt.Errorf("%s jenny failed on %s schema for %s: %w", j.inner.JennyName(), sfg.Schema.Version(), decl.Properties.Common().Name, err)
	}
	if f == nil || !f.Exists() {
		return nil, nil
	}

	f.RelativePath = filepath.Join(j.parentdir, comm.MachineName, f.RelativePath)
	f.From = append(f.From, j)
	return f, nil
}
