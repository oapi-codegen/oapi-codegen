package matrixv3allofintegernumber

// integer narrows number: Subject must be int. This fails to compile otherwise.
var _ Subject = int(3)
