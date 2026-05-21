# Building

This repository includes compiled files, so in most cases you can start running the server
as detailed in the execution chapter.
In case you need to rebuild, run make in the top level directory.
This builds all the needed components as well as the executable called main.

# Execution

Before starting the server you need to provide two files.
First is a date file, it's called updated.txt by default. The file should be provided with this repo.
You also need to provide a database file, this is not part of the repository. By default
the program looks for a database called neidb, you may change the file path using call arguments.
While starting up the server will report whether any of these files are missing.

You may run the executable without any arguments.
If you do so the server will be hosted at localhost:8080.

See usage for hosting remote.
