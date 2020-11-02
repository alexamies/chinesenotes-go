# Material Design Web Resources

This directory contains a set of Material Design Web HTML templates and related
resources for building a more usage web application that with the basic
HTML templates provided for the Quickstart.

## Setup

Clone this GitHub project, according to the main README.md.

Install Node.js.

## Building the 

Install the dependent Node packages:

```shell
npm install
```


Build the CSS and JavaScript bundles with the command

```shell
npm run-script build
```

This will generate transpiled bundles in the dist directory. Copy those to the
`web` directory where they can be served as static file by the go web app.

```shell
cp dist/* ../web/.
```

Edit the `config.yaml` file to set the TemplateDir app variable to this 
directory.

```
TemplateDir: web-resources 
```

Restart the web app.