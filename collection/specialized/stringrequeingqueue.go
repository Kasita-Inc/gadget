package specialized

import "github.com/Kasita-Inc/gadget/collection"

// NewStringRequeuingQueue that is empty and ready to use.
func NewStringRequeuingQueue() collection.StringStack {
	return collection.NewStringStackFromStack(NewRequeueingQueue())
}
