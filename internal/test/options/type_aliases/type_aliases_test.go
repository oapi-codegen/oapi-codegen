package optionstypealiases

// outputoptions/disabletypealiases: MustCompile will only compile if the type it is
// defined on has type aliases disabled. When disable-type-aliases-for-type includes
// "array", Example is emitted as `type Example []MyItem` (a named type), not
// `type Example = []MyItem` (a type alias), so pointer-receiver methods are valid.
func (*Example) MustCompile() {
}
