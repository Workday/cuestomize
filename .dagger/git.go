package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
)

type GitRepository struct {
	dir *dagger.Directory
}

func FromDirectory(dir *dagger.Directory) *GitRepository {
	return &GitRepository{
		dir: dir,
	}
}

func (r *GitRepository) Directory() *dagger.Directory {
	return r.dir
}

func (r *GitRepository) AsGit() *dagger.GitRepository {
	return r.Directory().AsGit()
}

func (r *GitRepository) Head() *dagger.GitRef {
	return r.AsGit().Head()
}

func (r *GitRepository) Stash(ctx context.Context) {
	r.Run(ctx, "stash")
}

func (r *GitRepository) Checkout(ctx context.Context, ref string) {
	r.Run(ctx, "checkout", ref)
}

func (r *GitRepository) output(c *dagger.Container) {
	r.dir = c.Directory("/git")
}

func (r *GitRepository) Run(ctx context.Context, args ...string) *dagger.Container {
	cmd := []string{"git"}
	cmd = append(cmd, args...)
	c := dag.Container().From(GitImage).
		WithWorkdir("/git").
		WithDirectory("/git", r.dir).
		WithExec(cmd)

	r.output(c)

	return c
}
