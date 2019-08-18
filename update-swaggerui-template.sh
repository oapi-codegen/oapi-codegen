#!/bin/bash

svn export https://github.com/swagger-api/swagger-ui/trunk/dist swagger-ui-dist

npx inliner swagger-ui-dist/index.html > swagger1.tmp

## Swagger UI
# npx prettier --parser html --write swagger1.tmp
sed 's,https://petstore.swagger.io/v2/swagger.json,__LEFT_DELIM__ .SpecURL __RIGHT_DELIM__,g' swagger1.tmp > swagger2.tmp
sed 's,/oauth2-redirect.html,__LEFT_DELIM__ .SwaggerUIRedirectURL __RIGHT_DELIM__,g' swagger2.tmp > swagger3.tmp
cp swagger3.tmp pkg/codegen/html/swagger.html

# Redirect page
# npx prettier --parser html --write swagger-ui-dist/oauth2-redirect.html
cp swagger-ui-dist/oauth2-redirect.html pkg/codegen/html/swagger-redirect.html

rm swagger1.tmp
rm swagger2.tmp
rm swagger3.tmp
rm swagger-redirect1.tmp
rm -r swagger-ui-dist
