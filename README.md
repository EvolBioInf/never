# Building

In case you need to rebuild, run `make` in the top level directory.
This builds all the needed components as well as the executables called `main` and `bin/never`.
`main` and `bin/never` are the same files, you may run the server using either of them.

# Execution

When starting the server you need to ensure the existence of a few files.
1) The date file. It's called `updated.txt` - a different name may be specified, when 
  running server. The file provides a timestamp of the last update. It should be part of this repo.
  This is an example of what the file might look like: Thu May 22 02:00:01 AM CEST 2025
2) Database directory. You need to provide a directory, which contains sqlite databases. We use
  `neighbors makeNeiDb` to create them. By default the server checks for a directory called `databases/`.
  You may change this name using the `-d` option.
3) Programs directory. Some endpoints rely on programs. These programs are part of the `neighbors`
  package. An executable for: `ants, dree, fintac, neighbors, ranks` and `taxi` need to provided in a
  directory called `prog`. This directory needs to exist in current working directory, when executing 
  the server. The name may not be changed when running the server.
4) Key and Cert. When running the server with HTTPS, you need to provide a `key.pem` and `cert.pem` file.
  You may change the names and file position using the `-k` and `-c` options.

When not providing any arguments the server runs locally using HTTP at port `8080`.
See usage for hosting remote.

# Testing

Run `make test` at the top level to execute all the tests. This creates two test databases, if not 
present and runs a testing server. The testing server runs at `http://localhost:8008` and is shutdown
after all the tests are done.
