package main

import (
	"context"
	"dagger/cuestomize/internal/dagger"
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

func (r *GitRepository) Checkout(ctx context.Context, ref string) error {
	_, err := r.Run(ctx, "checkout", ref).ExitCode(ctx)
	return err
}

func (r *GitRepository) Run(ctx context.Context, args ...string) *dagger.Container {
	cmd := []string{"git", "--git-dir=/git/state", "--work-tree=/git/worktree"}
	cmd = append(cmd, args...)
	return dag.Container().From(GitImage).
		WithWorkdir("/git/worktree").
		WithDirectory("/git/state", r.git).
		WithDirectory("/git/worktree", r.worktree).
		WithExec(cmd)
}
