package templater

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/google/go-github/v68/github"
	"github.com/sirupsen/logrus"
)

type wfParameters struct {
	Image               string
	BuildPre, BuildPost string
}

func (t *Templater) RenderWorkflows(ctx context.Context) error {
	for owner, ownerRepos := range t.config.Repositories {
		for repo, repoConfig := range ownerRepos {
			if err := t.renderWorkflows(ctx, owner, repo, repoConfig); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *Templater) loadWorkflowTemplates() {
	t.loadTemplates.Do(func() {
		const workflowsDir = "workflows"
		files, err := os.ReadDir(workflowsDir)
		if err != nil {
			panic(err)
		}

		t.templates = make(map[string]*template.Template)
		for _, f := range files {
			fp := filepath.Join(workflowsDir, f.Name())
			b, err := os.ReadFile(fp)
			if err != nil {
				panic(err)
			}
			fn := f.Name()
			t.templates[fn] = template.Must(template.New(fn).Delims("{{{", "}}}").Funcs(sprig.TxtFuncMap()).Parse(string(b)))
		}
	})
}

func (t *Templater) renderWorkflows(ctx context.Context, owner, repo string, cfg *RepositoryConfiguration) error {
	ref, _, err := t.client.Git.GetRef(ctx, owner, repo, "heads/main")
	if err != nil {
		return fmt.Errorf("getting base ref: %w", err)
	}
	refSHA := ref.GetObject().GetSHA()
	logrus.WithFields(logrus.Fields{
		"ref": refSHA,
	}).Info("loaded base ref")

	t.loadWorkflowTemplates()
	params := wfParameters{
		Image: fmt.Sprintf("%s/%s", t.config.Registry, repo),
	}
	if cfg != nil {
		params.BuildPre = cfg.PreBuild
		params.BuildPost = cfg.PostBuild
	}

	var entries []*github.TreeEntry
	for name, tmpl := range t.templates {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, params); err != nil {
			return err
		}
		rendered := buf.String()

		// Skip if the content matches the current file:
		wfPath := fmt.Sprintf(".github/workflows/%s", name)
		if existing, _, _, err := t.client.Repositories.GetContents(ctx, owner, repo, wfPath, &github.RepositoryContentGetOptions{Ref: refSHA}); err == nil {
			if ec, err := existing.GetContent(); err == nil && ec == rendered {
				continue
			}
		}

		entries = append(entries, &github.TreeEntry{
			Path:    github.Ptr(wfPath),
			Type:    github.Ptr("blob"),
			Content: github.Ptr(rendered),
			Mode:    github.Ptr("100644"),
		})
	}

	if len(entries) == 0 {
		return nil
	}

	tree, _, err := t.client.Git.CreateTree(ctx, owner, repo, refSHA, entries)
	if err != nil {
		return fmt.Errorf("creating tree: %w", err)
	}

	commit, _, err := t.client.Git.CreateCommit(ctx, owner, repo, &github.Commit{
		Message: github.Ptr("Update workflows"),
		Author:  t.config.Committer,
		Tree:    tree,
		Parents: []*github.Commit{
			{SHA: ref.GetObject().SHA},
		},
	}, &github.CreateCommitOptions{})
	if err != nil {
		return fmt.Errorf("creating commit: %w", err)
	}
	logrus.WithField("commit", commit.GetSHA()).Info("created commit")

	ref.Object.SHA = commit.SHA
	if _, _, err = t.client.Git.UpdateRef(ctx, owner, repo, ref, false); err != nil {
		return fmt.Errorf("updating ref: %w", err)
	}
	return nil
}
