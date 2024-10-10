# Dors

Dors is a tool to generate docs for your go project. It is based on the go doc tool and it generates a markdown file with the documentation of your project.

## Installation

```bash
go install github.com/ulm0/dors@latest
```

## Usage

```bash
$ docs gen --help
generate docs for your go project

Usage:
  dors gen [flags]

Flags:
  -e, --exclude-paths strings      A list of folders to exclude from the documentation.
  -h, --help                       help for gen
  -i, --include-sections strings   A list of sections to include in the documentation. (default [constants,factories,functions,methods,types,variables])
  -p, --print-source               Print source code for each symbol.
  -r, --recursive                  Read all files in the package and generate the documentation. It can be used in combination with include, and exclude. (default true)
  -c, --respect-case               Respect case when matching symbols. (default true)
  -s, --short                      One-line representation for each symbol.
  -k, --skip-sub-pkgs              SkipSubPackages will omit the sub packages section from the README.
  -t, --title string               Title for the documentation, if empty the package name is used.
  -u, --unexported                 Include unexported symbols.
```

This will generate a `DOCS.md` file in for each package in your project, processing the comments in your code.

---

## Acknowledgements

This is project is based on previous work made by [goreadme](https://github.com/posener/goreadme).