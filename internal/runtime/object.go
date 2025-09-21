package runtime

type ObjectType byte

const (
	OBJ_FUNCTION ObjectType = iota
	OBJ_NATIVE_FUNCTION
	OBJ_CLOSURE
	OBJ_UPVALUE
	OBJ_CLASS
	OBJ_INSTANCE
)

type Object interface {
	ObjectType() ObjectType

	String() string
	MarshalJSON() ([]byte, error)
}
type SerializableObject interface {
	Object
	serializable
}

// type ObjectList struct {
// 	objects *list.List
// }

// func NewObjectList() *ObjectList {
// 	return &ObjectList{
// 		objects: list.New(),
// 	}
// }

// func (ol *ObjectList) Add(obj *Object) *list.Element {
// 	return ol.objects.PushBack(obj)
// }

// func (ol *ObjectList) Remove(e *list.Element) {
// 	ol.objects.Remove(e)
// }

// func (ol *ObjectList) Clear() {
// 	ol.objects.Init()
// }
