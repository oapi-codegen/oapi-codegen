// Package parametersroundtrip is the multi-framework parameter binding roundtrip
// harness: one shared spec (parameters.yaml), one generated client, one table-driven
// test (param_roundtrip_test.go), and a thin per-framework server adapter + generated
// server under each framework subdirectory (chi, echo, echov5, fiber, gin, gorilla,
// iris, stdhttp). The client serializes Go values into a request, the server echoes
// them back as JSON, and the test asserts the roundtrip for every parameter style.
//
// Folds in: parameters/{chi,client,echo,echov5,fiber,gin,gorilla,iris,stdhttp}
package parametersroundtrip
