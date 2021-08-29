# actions-templates

Reads a [SOPS](https://github.com/mozilla/sops) configuration file and offers two features:
* `secrets` - synchronize secrets with each repository
* `workflows` - render workflow templates from `workflows/` directory


## Configuration

```yaml
repositories:
    my-org:
        # Configuration is optional
        my-awesome-repo: null
        my-cool-repo:
            prebuild: |-
                - name: Load from cache
                  run: actions/cache@v2
committer:
    name: workflow updater
    email: workflows@updater.com
auth:
    github: ghp_lolololol
secrets:
    MY_REGISTRY_PASSWORD: bar
```

### Sample workflow template

```yaml
name: Build
on:
  push:
permissions:
  contents: read
jobs:
  test:
    runs-on: self-hosted
    steps:
    - name: "🤓 Fetching code"
      uses: actions/checkout@5a4ac9002d0be2fb38bd78e4b4dbde5606d7042f
    - uses: docker/login-action@f054a8b539a109f9f41c372932f1ae047eff08c9
      with:
        registry: {{{splitList "/" .Image | first}}}
        username: my-username
        password: ${{secrets.MY_REGISTRY_PASSWORD}}
{{{.BuildPre | indent 4}}}
    - name: "🚧 Build image"
      run: docker build --cache-from "{{{.Image}}}:latest" -t "{{{.Image}}}:${{github.sha}}" .
{{{.BuildPost | indent 4}}}
```
