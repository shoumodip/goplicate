# Goplicate
Gather duplicate files across the filesystem

## Quick Start
```console
$ go build goplicate.go
```

## Usage
```console
$ ./goplicate gather extra  # Create folder 'extra' and move duplicate files into it
$ ./goplicate restore extra # Move files from folder 'extra' to their original locations
```
