package operation

import (
	"fmt"

	"github.com/hackerwins/yorkie/pkg/document/json"
	"github.com/hackerwins/yorkie/pkg/document/json/datatype"
	"github.com/hackerwins/yorkie/pkg/document/time"
	"github.com/hackerwins/yorkie/pkg/log"
)

type Add struct {
	value           datatype.Element
	parentCreatedAt *time.Ticket
	prevCreatedAt   *time.Ticket
	executedAt      *time.Ticket
}

func NewAdd(
	value datatype.Element,
	parentCreatedAt *time.Ticket,
	prevCreatedAt *time.Ticket,
	executedAt *time.Ticket,
) *Add {
	return &Add{
		value:           value,
		parentCreatedAt: parentCreatedAt,
		prevCreatedAt:   prevCreatedAt,
		executedAt:      executedAt,
	}
}

func (o *Add) Execute(root *json.Root) error {
	parent := root.FindByCreatedAt(o.parentCreatedAt)

	obj, ok := parent.(*json.Array)
	if !ok {
		err := fmt.Errorf("fail to execute, only Array can execute Set")
		log.Logger.Error(err)
		return err
	}

	obj.InsertAfter(o.prevCreatedAt, o.value)
	root.RegisterElement(o.value)
	return nil
}

func (o *Add) Value() datatype.Element {
	return o.value
}

func (o *Add) ParentCreatedAt() *time.Ticket {
	return o.parentCreatedAt
}

func (o *Add) ExecutedAt() *time.Ticket {
	return o.executedAt
}

func (o *Add) SetActor(actorID *time.ActorID) {
	o.executedAt = o.executedAt.SetActorID(actorID)
}

func (o *Add) PrevCreatedAt() *time.Ticket {
	return o.prevCreatedAt
}
