# AXIS VFS, a simple virtual file system API.

AXIS is based on a few simple interfaces and a set of API functions that operate on these interfaces.
Clients use the provided implementations of these interfaces (or provide their own custom implementations)
to create "data sources" that may be mounted on a "file system" and used for OS-independent file IO.

AXIS was originally written to allow files inside of archives to be handled with exactly the same API as
used for files inside of directories, but it has since grown to allow "logical" files and directories as
well as "multiplexing" multiple items on the same location (to, for example, make two directories look
and act like one). These properties make AXIS perfect for handling data and configuration files for any
program where flexibility is important, the program does not need to know where its files are actually
located, it simply needs them to be at a certain place in it's AXIS file system. Changing where a program
loads it's files from is then as simple as changing the code that initializes the file system.

AXIS uses standard slash separated paths. File names may not contain any "special" characters, basically
any of the following:

	< > ? * | : " \ /

Additionally "." and ".." are not valid names.

Multiple slashes in an AXIS path are condensed into a single slash, and any trailing slashes will be stripped off.
For example the following two paths are equivalent:

	test/path/to/dir
  /test///path//to/dir/

Obviously you should always use the first form, but the second is still legal (barely)

AXIS VFS "officially" stands for Absurdly eXtremely Incredibly Simple Virtual File System (adjectives are
good for making cool acronyms!). If you think the name is stupid (it is) you can just call it AXIS and
forget what it is supposed to mean, after all the "official" name is more of a joke than anything...
