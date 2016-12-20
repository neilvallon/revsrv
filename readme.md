# revsrv
A simple static file server for multiple domains.

## folder structure
Starting at the working directory revsrv will look for files in reverse domain name notation followed by the requested file path. e.g. `www.example.com/folder/index.html` would request the file `com/example/www/folder/index.html`.

A basic layout for a server with 3 domains might look something like this:

- public
	- com
		- example
			- www
				- index.html
		- example2
			- errors
				- 404.html
	- net
		- example3
			- ...

One quirk of this is that `www.example.com/...` will be equivelant to `example.com/www/...`, but browsers will treat absolute file references on the page differently.

## alias.csv format
There are a few rewrite rules that are possible to subvert the basic structure of a revsrv server.

- `dummydomain.com,realdomain.com`
	* Used to alias domains.
	* In order to avoid loops, aliases are not transitive and will not redirect browsers.
	* References to the first domain simply point to the files of the second.


- `*,mainsite.com`
	* Treats any request not matched as if it were a master domain.

- `[404],/errors/404.html`
	* Attempts to use the provided page path whenever the given error occurs.
	* Paths are followed relative to the domain folder.
	* Each domain can have its own error pages, but must follow a consistant layout.
