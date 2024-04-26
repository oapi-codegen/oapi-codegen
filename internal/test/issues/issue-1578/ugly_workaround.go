package issue1578

// This is an ugly workaround for the issue
//var _ json.Marshaler = Test200JSONResponse{}
//
//func (response Test200JSONResponse) MarshalJSON() ([]byte, error) {
//	return json.Marshal(ComplexObject(response))
//}
