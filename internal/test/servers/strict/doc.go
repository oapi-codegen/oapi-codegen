// Package serversstrict is the multi-framework strict-server roundtrip harness: one
// shared spec (strict-schema.yaml), one generated client, one table-driven test
// (strict_test.go), and a thin per-framework strict server adapter + generated server
// under each framework subdirectory (chi, echo, fiber, gin, gorilla, iris, stdhttp).
//
// Folds in: strict-server/{chi,client,echo,fiber,gin,gorilla,iris,stdhttp}
// (issue-1529 multi-framework strict and issue-1963 response buffering / error handler
// are folded as additional cases — see below).
package serversstrict
