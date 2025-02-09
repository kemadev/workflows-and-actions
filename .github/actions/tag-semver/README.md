# `.github/actions/tag-semver`

## Files in this directory

- Defines a GitHub Action that checks tags the next semantic version of the repository

## Usage

```yaml
  - name: Tag version
    id: tag
    uses: kemadev/workflows-and-actions/.github/actions/tag-semver@main
    env:
      GH_TOKEN: ${{ github.token }}
      FORCE_PATCH: <boolean>
```
