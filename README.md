# K9s_help

Welcome to the K9s Helper! K9s is a powerful command-line tool for managing Kubernetes clusters. It offers a range of features to help you quickly back up and migrate your Kubernetes resources.

- Cluster Backup
- Quick Migration
- Command-Line
- Rich Functionality

# Dependency requirements

Minimum Supported Golang Version is 1.20.


# Getting started

**Download package**
[latest version 1.0.1](https://github.com/xxscloud5722/k9s-help/releases)

**Program compilation**
- Golang 1.20

```bash
# windows OR linux
go env -w GOOS=linux
go mod tidy
go build -o ./dist/k9s_help src/main.go
```

# K9sHelp Guide

Please place the program in the global path and set up the environment variables.

The panel machine has kubectl installed, enabling access to additional functionalities.

```bash
# Default /root/.kube/conf

# Kubernetes YAML synchronization
k9s_help 
```

# Contributors

Thanks for your contributions!

- [@xiaoliuya](https://github.com/xxscloud5722/)

# Zen
Don't let what you cannot do interfere with what you can do.
