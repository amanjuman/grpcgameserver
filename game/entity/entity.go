package entity

import (
	"fmt"
	. "github.com/amanjuman/grpcgameserver/message"
	"github.com/amanjuman/grpcgameserver/physic"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

type Entity struct {
	sync.RWMutex
	EntityInfo *Character
	TypeName   string
	Health     float32
	Alive      bool
	I          IEntity
	GM         *GameManager
	Room       *Room
	World      *physic.World
	Obj        *physic.Obj
	Skill      map[string]AttackBehavier
}
type IEntity interface {
	IGameBehavier
	Hit(int32)
	GetInfo() *Character
	Init(*GameManager, *Room, *Character)
	Move(in *Input)
	GetTransform() *TransForm
	Harm(blood float32)
}

func (e *Entity) GetInfo() *Character {
	e.RLock()
	entityInfo := proto.Clone(e.EntityInfo).(*Character)
	e.RUnlock()
	return entityInfo
}
func (e *Entity) Hit(damage int32) {
	fmt.Println("-", damage)
}

func (e *Entity) Harm(blood float32) {
	e.Lock()
	e.Health -= blood
	if e.Health <= 0 {
		//Dead
		e.Alive = false
		e.Unlock()
		e.Destroy()
		return
	}
	f := &CallFuncInfo{
		Func:     "Health",
		Value:    e.Health,
		TargetId: e.EntityInfo.Uuid,
	}
	e.Room.SendFuncToAll(f)
	e.Unlock()
}

func (e *Entity) Init(gm *GameManager, room *Room, entityInfo *Character) {
	e.GM = gm
	e.EntityInfo = entityInfo
	e.Room = room
	e.World = room.World
	var ok bool
	e.Obj, ok = room.World.Objs.Get(entityInfo.Uuid)
	if !ok {
		log.Fatal("[entity]{init} Get obj ", entityInfo.Uuid, " is not found. ")
	}
	//call All client create enitity at some point
	e.costumeInit()
}

func (e *Entity) costumeInit() {
	log.Warn("Please define your costumeInit")
}
func (e *Entity) Tick() {
}
func (e *Entity) Destroy() {
	e.Lock()
	e.GM.DestroyEntity(e.EntityInfo.Uuid)
	e.Room.DestroyEntity(e.EntityInfo.Uuid)
	e.World.DeleteObj(e.EntityInfo.Uuid)
	e.Obj.Destroy()
	e.Obj = nil
	e.Unlock()
	e = nil
}

func (e *Entity) Run() {
}
func (e *Entity) PhysicUpdate() {
}

func (e *Entity) GetTransform() *TransForm {
	return &TransForm{}
}
func (e *Entity) Move(in *Input) {
	turnSpeed := e.EntityInfo.Ability.TSPD
	moveSpeed := e.EntityInfo.Ability.SPD
	moveValue := in.V_Movement
	turnValue := in.H_Movement
	e.Room.World.Move(e.EntityInfo.Uuid, float64(moveValue*moveSpeed), float64(turnValue*turnSpeed))
}
