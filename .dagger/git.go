package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
)

const (
	stateDir    = "/git/state"
	worktreeDir = "/git/worktree"
)

type GitRepository struct {
	git      *dagger.Directory
	worktree *dagger.Directory
}

func FromDirectory(dir *dagger.Directory) *GitRepository {
	return &GitRepository{
		git:      dir.Directory(".git"),
		worktree: dir.WithoutDirectory(".git"),
	}
}

func (r *GitRepository) Directory() *dagger.Directory {
	return r.worktree.WithDirectory(".git", r.git)
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
	r.git = dag.Directory().WithDirectory(".git", c.Directory(stateDir))
	r.worktree = dag.Directory().WithDirectory(".", c.Directory(worktreeDir))
}

func (r *GitRepository) Run(ctx context.Context, args ...string) *dagger.Container {
	cmd := []string{"git", "--git-dir=" + stateDir, "--work-tree=" + worktreeDir}
	cmd = append(cmd, args...)
	c := dag.Container().From(GitImage).
		WithWorkdir(worktreeDir).
		WithDirectory(stateDir, r.git).
		WithDirectory(worktreeDir, r.worktree).
		WithExec(cmd)

	r.output(c)

	return c
}
