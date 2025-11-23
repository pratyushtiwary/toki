package toki

func Check(e error, prefix string) {
	if e != nil {
		panic(e)
	}
}
