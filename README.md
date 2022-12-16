# OpenAPI Client and Server Code Generator

This repo is a fork of [deepmap/oapi-codegen](https://github.com/do87/oapi-codegen)

Only the changes from the original repo are documented below

## Spec tidying

In cases where the OpenAPI spec isn't created by whomever generates the client, a "tidy" functionality is used to make the client more readable / styled as needed

In a configuration file, the following code can be added:

```yaml
tidy:
  verbose: false
  functions:
  - replace: service_
    with: 
    prefix: true
  params:
  - replace: Id
    with: ID
    suffix: true
  schemas:
  - replace: Cpu
    with: CPU
    match: true

```

## Splitting client code by tags

the config has been extended to support splitting the client code into multiple directories and files

Example:

```yaml
output-options:
  split-by-tags:
    verbose: true
    enabled: true
```

theres also an options to add a list under `split-by-tags` of `includes` or `excludes`

## Extend responses struct

For our use case we'd like to add an aggregated error if the response isn't 200

In order to do that the response struct can be extended with the following config

Example:

```yaml
output-options:
  extend-response:
  - field: HasError
    type: error
    description: "Aggregated error"
    apply-to: ["*"]
    imports: ["errors"]
```
