#!/bin/bash

svn export https://github.com/swagger-api/swagger-ui/trunk/dist swagger-ui-dist

npx inliner swagger-ui-dist/index.html > swagger1.tmp
# npx prettier --parser html --write swagger1.tmp
sed 's,https://petstore.swagger.io/v2/swagger.json,__LEFT_DELIM__.__RIGHT_DELIM__,g' swagger1.tmp > swagger2.tmp

cp swagger2.tmp pkg/codegen/templates/swagger.html.tmpl

rm swagger1.tmp
rm swagger2.tmp
rm -r swagger-ui-dist
