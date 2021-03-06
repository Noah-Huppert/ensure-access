# Ensure Access
A tool which ensures that a certain level of access exists for files 
and directories.

# Table Of Contents
- [Overview](#overview)
- [Usage](#usage)
- [Build](#build)

# Overview
Makes sure that files and directories have the required permissions.

# Usage
```
ensure-access -path PATH -mode MODE_OCTALS -poll POLL_SECS
```

The `-path PATH` argument can be specified more than once.  

Run `ensure-access -h` for more information.

# Build
Build:

```
make build
# or 
make
```
