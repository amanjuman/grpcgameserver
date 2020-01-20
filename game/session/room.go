package session

import (
	"errors"
	. "github.com/amanjuman/grpcgameserver/message"
	. "github.com/amanjuman/grpcgameserver/uuid"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type ChatMessage struct {
	SpeakerId   int32
	SpeakerName string
	Content     string
}

type ChatRoom struct {
	ReadingBuffer []*ChatMessage
}

var limitReadyTime = 1000

var RoomManager *roomManager = &roomManager{
	Rooms:                make(map[int64]*Room),
	UserIdleRoomListChan: make(map[*MsgChannel]struct{}),
}

type roomManager struct {
	UserIdleRoomListChan map[*MsgChannel]struct{}
	Rooms                map[int64]*Room
	sync.RWMutex
	RoomList *RoomList
}

func (rm *roomManager) CreateGame(gameCreation *GameCreation) error {
	switch gameCreation.RoomInfo.GameType {
	default:
		if _, ok := rm.Rooms[gameCreation.RoomInfo.Uuid]; ok {
			return errors.New("Room already exist")
		}
		room := NewRoom(gameCreation.RoomInfo.Uuid)
		for _, sessionInfo := range gameCreation.PlayerSessions {
			session := Manager.CreateSessionFromAgent(sessionInfo)
			log.Debug("[CreateGame]", session)
			session.InputPool = room.GetMsgChan("Input")
			room.Client[session] = struct{}{}
			session.Room = room

		}
		rm.Rooms[gameCreation.RoomInfo.Uuid] = room
		go room.Run()
		return nil
	}
}

func (rm *roomManager) DeleteRoom(info *RoomInfo) error {
	room, ok := rm.Rooms[info.Uuid]
	if !ok {
		log.Debug(rm.Rooms)
		return errors.New("no this room")
	}
	room.end <- struct{}{}
	delete(rm.Rooms, info.Uuid)
	log.Debug("Game Rooms After delete", rm.Rooms)
	return nil
}

func (rm *roomManager) LeaveRoom() {

}

func (rm *roomManager) CreateRoom(master *Session, setting *RoomSetting) *Room {
	id, _ := Uid.NewId(ROOM_ID)
	room := NewRoom(id)
	rm.Rooms[id] = room
	return room
}

func NewRoom(roomId int64) *Room {
	room := &Room{
		Client:            make(map[*Session]struct{}),
		Uuid:              roomId,
		GameStart:         make(chan (struct{}), 1),
		MsgChannelManager: NewMsgChannelManager(),
		end:               make(chan struct{}, 10),
	}
	room.AddMsgChan("Input", 200)
	return room
}

type Room struct {
	Name      string
	GameType  string
	Master    *Session
	Client    map[*Session]struct{}
	Uuid      int64
	GameFrame *GameFrame
	*MsgChannelManager
	sync.RWMutex
	GameStart chan (struct{})
	end       chan (struct{})
}

func (r *Room) GenerateStartFrame() {
	r.GameFrame.EntityStates = make(map[int64]*EntityState)
	r.GameFrame.Interaction = make([]*Interaction, 0)
	r.GameFrame.Characters = make(map[int64]*Character)
	for s, _ := range r.Client {

		c := s.Info.UserInfo.OwnCharacter[s.Info.UserInfo.UsedCharacter]
		r.GameFrame.EntityStates[s.Info.UserInfo.UsedCharacter] = NewEntityState(s.Info.UserInfo.UsedCharacter, "Tank", c)
		r.GameFrame.Characters[s.Info.UserInfo.UsedCharacter] = c
		log.Debug(c, s.Info.UserInfo.UsedCharacter)
	}
	r.SyncGameFrame()
	//r.GameFrame.Characters = make(map[int64]*Character)
}

func NewEntityState(id int64, prefabName string, c *Character) *EntityState {
	es := &EntityState{
		Health:     c.MaxHealth,
		Uuid:       id,
		PrefabName: prefabName,
		Transform: &Transform{
			Position: &Vector3{0, 0, 0},
			Rotation: &Quaternion{0, 0, 0, 1},
		},
		Animation: &Animation{},
		Speed:     &Vector3{},
	}
	return es
}

func (r *Room) Run() {
	//Syncpos
	inputPool := r.GetMsgChan("Input")
	for _, _ = range r.Client {
		<-r.GameStart
	}
	r.GameFrame = &GameFrame{}
	r.GameFrame.RunnigNo = 0
	r.GameFrame.TimeStamp = GetTimeStamp()
	r.GenerateStartFrame()
	r.GameFrame.RunnigNo += 1
	update := time.NewTicker(time.Millisecond * 30)
END:
	for {
		select {
		case <-update.C:
			//log.Debug("GameFrame: ", r.GameFrame)
			r.UpdateFrame()
		case msg := <-inputPool.DataCh:
			input := msg.(*Input)
			r.HandleEntityState(input)
			r.HandleNewEntity(input)
			r.HandleInteraction(input)
			r.HandleEntityDestory(input)
		case <-r.end:
			break END
		}
	}

	log.Debug("End Gameing")
	for s, _ := range r.Client {
		s.Room = nil
	}
	r.Master.Room = nil
}

func (r *Room) UpdateFrame() {
	r.SyncGameFrame()
	r.GameFrame.RunnigNo += 1
	r.GameFrame.TimeStamp = GetTimeStamp()
	r.GameFrame.Interaction = make([]*Interaction, 0)
}

func (r *Room) HandleEntityDestory(input *Input) {
	if len(input.DestroyEntity) > 0 {
		//r.UpdateFrame()
		for _, id := range input.DestroyEntity {
			delete(r.GameFrame.Characters, id)
			delete(r.GameFrame.EntityStates, id)
		}
	}
}

func (r *Room) HandleInteraction(input *Input) {
	for _, in := range input.Interaction {
		r.GameFrame.Interaction = append(r.GameFrame.Interaction, in)
	}
}

func (r *Room) HandleEntityState(input *Input) {
	for _, in := range input.EntityStates {
		r.GameFrame.EntityStates[in.Uuid] = in
	}
}

func (r *Room) HandleNewEntity(input *Input) {
	for id, in := range input.NewEntityCharacters {
		r.GameFrame.Characters[id] = in
	}
}

func (r *Room) SyncGameFrame() {
	gf := proto.Clone(r.GameFrame).(*GameFrame)
	for s, _ := range r.Client {
		msg := s.GetMsgChan("GameFrame")
		msg.DataCh <- gf
	}
}

func (r *Room) PlayerLeave(s *Session) {
	delete(r.Client, s)
	if len(r.Client) == 0 {
		r.end <- struct{}{}
	}
}

func GetTimeStamp() int64 {
	return int64(time.Now().UnixNano() / 1000000)
}
