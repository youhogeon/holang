package runtime

import "container/list"

type ObjectType byte

const (
	OBJ_FUNCTION ObjectType = iota
)

// type Object struct {
// 	ObjectType ObjectType
// 	Value      any
// }

type Object interface {
	ObjectType() ObjectType
}

type ObjectList struct {
	objects *list.List
}

func NewObjectList() *ObjectList {
	return &ObjectList{
		objects: list.New(),
	}
}

func (ol *ObjectList) Add(obj *Object) *list.Element {
	return ol.objects.PushBack(obj)
}

func (ol *ObjectList) Remove(e *list.Element) {
	ol.objects.Remove(e)
}

func (ol *ObjectList) Clear() {
	ol.objects.Init()
}
