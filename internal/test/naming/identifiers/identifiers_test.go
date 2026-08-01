package namingidentifiers

// Compile-only tests: all three folded sources had no _test.go files.
// The mere fact that identifiers.gen.go compiles verifies:
//
//   issue-head-digit-of-operation-id: operationId "3GPPFoo" (leading digit)
//     is emitted as a valid Go identifier in the StrictServerInterface.
//
//   issue-head-digit-of-httpheader: header name "000-foo" (leading digit)
//     is emitted as field N000Foo in the response-headers struct.
//
//   issue1767: struct field "_id" (leading underscore) is emitted as
//     UnderscoreId in the Alarm struct.
