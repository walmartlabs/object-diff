package obj_diff

type structInt32 struct {
	A int32
}

type structStruct struct {
	A structInt32
}

type structMap struct {
	A map[string]int32
}

type structArray3 struct {
	A [3]int32
}

type structSlice struct {
	A []int32
}

type structPtr struct {
	A *int32
}
