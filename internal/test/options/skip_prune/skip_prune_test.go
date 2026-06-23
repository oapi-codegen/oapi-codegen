package optionsskipprune

// issue #240: compile-only — the source had no runtime assertions.
//
// The contrast between default-prune and skip-prune is captured structurally:
//
//   default_prune.gen.go  — generated with skip-prune unset; the Unreferenced
//                           schema is absent from the output (pruned).
//
//   skip_prune.gen.go     — generated with skip-prune:true; the Unreferenced
//                           type is present and this file must compile.
//
// If either generation run fails or the package fails to compile, the test
// fails automatically (go test requires the package to build).
