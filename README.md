# OpenAPI Client and Server Code Generator

This repo is a fork of [deepmap/oapi-codegen](https://github.com/deepmap/oapi-codegen)

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
