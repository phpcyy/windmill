package generator

import "testing"

func TestBuilder_Build(t *testing.T) {
	b := NewBuilder("../schemes/entity")
	err := b.Build("blog.yml")
	if err != nil {
		t.Error(err)
		return
	}
	err = b.FlushRoutes()
	if err != nil {
		t.Error(err)
		return
	}
}
