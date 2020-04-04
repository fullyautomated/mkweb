![Fully Automated Technologies Logo](https://fully.automated.ee/img/fa-banner.svg)
# mkweb static site generator

mkweb is a simple static site generator, born out of the need to just have something really easy to use and hassle-free. Most other options like, e.g. hugo require a non-trivial amount of work to convert exisiting html code or to write a template from scratch.

## Usage

To convert a single file:

`./mkweb -file index.md`

To serve a directory:

`./mkweb -path template/`

When serving a directory, it also sends filesystem events via WebSockets to the currently open webpages, which can then be used to reload on save. For an example as to how this feature is used, see the template/ folder.
