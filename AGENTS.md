# undercov

## Data structure

The coverage files are stored in a separate branch in the same Git repository. We support monorepos with multiple coverage files. The structure of the branch is as follows:

```
- .undercov/ <- This is the root directory for undercov data. The name is hardcoded and should not be changed.
  - [...branch-name/]/ <- This is the directory for each branch. If the branch name has slashes, it leads to nested directories. For example, if the branch name is `feature/coverage`, the path would be `.undercov/feature/coverage/`.
    - [base64-encoded-file-path].lcov <- Each coverage file is stored with its path encoded in base64, so multiple files can be stored without conflicts. The file extension is `.lcov` to indicate that it is a coverage file.
```
